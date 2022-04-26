import { isArray, chunk } from "lodash";
const fetch = require("node-fetch");
import { S3, DynamoDB, SQS } from "aws-sdk";
import { addDays, format, parse, subDays } from "date-fns";

export type FuelType = "Unleaded95" | "Octane100" | "Diesel";
function fuelTypeToOkItemNumber(fuelType: FuelType): number {
  switch (fuelType) {
    case "Unleaded95":
      return 536;
    case "Octane100":
      return 533;
    case "Diesel":
      return 231;
    default:
      return 536;
  }
}

export interface OkPrices {
  historik: {
    dato: string;
    pris: number;
    prisExclAfgifterExclMoms?: number;
    prisExclAfgifterInclMoms?: number;
    prisExclMoms?: number;
    varenr?: number;
  }[];
}

export interface Price {
  date: string;
  price: number;
}

export interface DayPrices {
  today: Price;
  yesterday: Price | null;
  tomorrow: Price | null;
}

// https://stackoverflow.com/a/15289883
function dateDiffInDays(a: Date, b: Date) {
  const _MS_PER_DAY = 1000 * 60 * 60 * 24;
  const utc1 = Date.UTC(a.getFullYear(), a.getMonth(), a.getDate());
  const utc2 = Date.UTC(b.getFullYear(), b.getMonth(), b.getDate());

  return Math.floor((utc2 - utc1) / _MS_PER_DAY);
}

export function createPriceService(): IPriceService {
  const bucketName = process.env.BUCKET || "";
  const tableName = process.env.TABLE_NAME || "";
  const sqsUrl = process.env.SQS_URL || "";
  console.log("NODE_ENV -> ", process.env.NODE_ENV);
  console.log("BUCKET ->", bucketName);
  console.log("TABLE -> ", tableName);
  console.log("SQSURL -> ", sqsUrl);

  const localhostEndpoint =
    process.env.NODE_ENV === "dev"
      ? "http://host.docker.internal:4566"
      : undefined;

  const s3 = new S3({
    endpoint: localhostEndpoint,
    s3ForcePathStyle: true,
  });

  const db = new DynamoDB.DocumentClient({
    endpoint: localhostEndpoint,
  });

  const sqs = new SQS({
    endpoint: localhostEndpoint,
  });

  return new PriceService(s3, bucketName, db, tableName, sqs, sqsUrl);
}

export interface IPriceService {
  getPrices: (date: Date, fuelType: FuelType) => Promise<DayPrices | null>;
  updatePriceCache: (fueltype: FuelType) => Promise<void>;
  doCacheWrite: (
    fueltype: FuelType,
    priceChunk: OkPrices["historik"]
  ) => Promise<void>;
  fetchPrices: (fuelType: FuelType) => Promise<void>;
}
export class PriceService implements IPriceService {
  private readonly s3: S3;
  private readonly s3Bucket: string;

  private readonly db: DynamoDB.DocumentClient;
  private readonly tableName: string;

  private readonly sqs: SQS;
  private readonly sqsUrl: string;

  constructor(
    s3: S3,
    s3Bucket: string,
    db: DynamoDB.DocumentClient,
    tableName: string,
    sqs: SQS,
    sqsUrl: string
  ) {
    this.s3 = s3;
    this.s3Bucket = s3Bucket;
    this.db = db;
    this.tableName = tableName;

    this.sqs = sqs;
    this.sqsUrl = sqsUrl;
  }
  /**
   * Get price for the date requested, the day before, and the day after if possible
   */
  async getPrices(date: Date, fuelType: FuelType): Promise<DayPrices | null> {
    const fuelprices: OkPrices = { historik: [] };
    const dateFormat = "yyyy-MM-ddT00:00:00";
    try {
      console.log("getting data from dynamodb");
      const yesterdayDate = subDays(date, 1);
      const tomorrowDate = addDays(date, 2);
      const result = await this.db
        .query({
          TableName: this.tableName,
          KeyConditionExpression: "PK = :PK AND SK BETWEEN :from AND :to",
          ExpressionAttributeValues: {
            ":PK": `FUELTYPE#${fuelType}`,
            ":from": `DATE#${format(yesterdayDate, dateFormat)}`,
            ":to": `DATE#${format(tomorrowDate, dateFormat)}`,
          },
        })
        .promise();
      (result.Items ?? []).forEach((_item) => {
        const item = _item as { SK: string; price: number; PK: string };

        const date = item.SK.split("#")[1];

        fuelprices.historik.push({
          dato: date,
          pris: item.price,
        });
      });
    } catch (error) {
      console.log("error getting dynamodb data", error);
    }

    function findPrice(
      historik: OkPrices["historik"],
      date: Date
    ): OkPrices["historik"][0] | null {
      const price = historik.find((price) => {
        const [year, month, day] = price.dato
          .split("T")[0]
          .split("-")
          .map((x) => Number(x));
        return (
          year === date.getFullYear() &&
          month === date.getMonth() + 1 &&
          day === date.getDate()
        );
      });
      if (!price) {
        return null;
      }
      return price;
    }

    const todayOkPrice = findPrice(fuelprices.historik, date);
    if (!todayOkPrice) {
      return null;
    }
    const prices: DayPrices = {
      today: {
        date: todayOkPrice.dato,
        price: todayOkPrice.pris,
      },
      tomorrow: null,
      yesterday: null,
    };

    const yesterdayDate = subDays(date, 1);
    const yesterdayOkPrice = findPrice(fuelprices.historik, yesterdayDate);
    if (yesterdayOkPrice) {
      prices.yesterday = {
        date: yesterdayOkPrice.dato,
        price: yesterdayOkPrice.pris,
      };
    }

    const tomorrowDate = addDays(date, 1);
    const tomorrowOkPrice = findPrice(fuelprices.historik, tomorrowDate);
    if (tomorrowOkPrice) {
      prices.tomorrow = {
        date: tomorrowOkPrice.dato,
        price: tomorrowOkPrice.pris,
      };
    }

    return prices;
  }

  async updatePriceCache(fueltype: FuelType): Promise<void> {
    const data = (
      await this.s3
        .getObject({
          Bucket: this.s3Bucket,
          Key: `prices/${fueltype}`,
        })
        .promise()
    )?.Body?.toString("utf-8");

    let priceResp: OkPrices | null = null;
    try {
      priceResp = JSON.parse(data || "") as OkPrices;
    } catch (error) {
      console.log("could not parse s3 data", error);
      return;
    }

    if (!priceResp) {
      console.log("priceResp was null");
      return;
    }

    const result = await this.db
      .query({
        TableName: this.tableName,
        KeyConditionExpression: "PK = :PK",
        ExpressionAttributeValues: {
          ":PK": `FUELTYPE#${fueltype}`,
        },
        Limit: 1,
        ScanIndexForward: false,
      })
      .promise();
    const firstDbItem = result.Items?.[0] as {
      SK: string;
      price: number;
      PK: string;
    };
    console.log("firstDbItem", firstDbItem?.SK);
    console.log(
      "firstPriceResp",
      priceResp.historik[priceResp.historik.length - 1].dato
    );
    const firstDbItemDate = firstDbItem?.SK.split("#")[1];

    // Nothing to do if first item is equal
    if (
      firstDbItemDate ===
        priceResp.historik[priceResp.historik.length - 1].dato &&
      firstDbItem?.price ===
        priceResp.historik[priceResp.historik.length - 1].pris
    ) {
      return;
    }

    // Loop through S3 data until we find an item with matching date
    const toInsert: OkPrices = {
      historik: [],
    };
    for (const item of priceResp.historik.reverse()) {
      toInsert.historik.push(item);
      if (item.dato === firstDbItemDate) {
        break;
      }
    }

    console.log("toInsert length", toInsert.historik.length);

    if (toInsert.historik.length === 0) {
      return;
    }

    const priceChunks = chunk(toInsert.historik, 25);

    const priceChunkBatches = chunk(priceChunks, 10);
    priceChunkBatches.forEach(async (batch, i) => {
      const entries: SQS.SendMessageBatchRequestEntryList = batch.map(
        (priceChunk, j) => {
          return {
            Id: `${i}-${j}`,
            MessageBody: JSON.stringify({
              fueltype,
              priceChunk,
            }),
          };
        }
      );
      await this.sqs
        .sendMessageBatch({
          Entries: entries,
          QueueUrl: this.sqsUrl,
        })
        .promise();
    });
  }

  async doCacheWrite(fueltype: FuelType, priceChunk: OkPrices["historik"]) {
    const ddbItems = priceChunk.map((price) => {
      return {
        PutRequest: {
          Item: {
            PK: `FUELTYPE#${fueltype}`,
            SK: `DATE#${price.dato}`,
            price: price.pris,
          },
        },
      };
    });

    console.log("doCacheWrite: ddbItems.length", ddbItems.length);
    if (ddbItems.length > 0) {
      console.log("doCacheWrite: ddbItems[0]", ddbItems[0]);
    }

    try {
      const resp = await this.db
        .batchWrite({
          RequestItems: {
            [this.tableName]: ddbItems,
          },
        })
        .promise();
    } catch (error) {
      console.error(`error writing chunk to dynamodb`, error);
    }
  }

  async fetchPrices(fuelType: FuelType): Promise<void> {
    let priceResp: OkPrices | null = null;
    try {
      const resp = await fetch(
        //'https://www.ok.dk/privat/produkter/benzinkort/prisudvikling/getProduktHistorik',
        "https://www.ok.dk/privat/produkter/ok-kort/prisudvikling/getProduktHistorik",
        {
          method: "POST",
          body: JSON.stringify({
            varenr: fuelTypeToOkItemNumber(fuelType),
            pumpepris: true,
          }),
          headers: {
            "Content-Type": "application/json",
          },
        }
      );
      priceResp = (await resp.json()) as OkPrices;
      priceResp.historik.forEach((price) => {
        delete price.prisExclAfgifterExclMoms;
        delete price.prisExclAfgifterInclMoms;
        delete price.prisExclMoms;
        delete price.varenr;
      });
    } catch (error) {
      console.error("error fetching data from OK", error);
    }

    if (!priceResp) {
      console.error("priceResp was null");
      return;
    }

    try {
      await this.s3
        .upload({
          Bucket: this.s3Bucket.toLowerCase(),
          Key: `prices/${fuelType}`,
          Body: JSON.stringify(priceResp),
        })
        .promise();
    } catch (error) {
      console.error("error uploading to S3", error);
    }
  }
}
