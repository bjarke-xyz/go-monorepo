import { isArray } from "lodash";
const fetch = require("node-fetch");
import { S3, DynamoDB } from "aws-sdk";

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
  console.log("NODE_ENV -> ", process.env.NODE_ENV);
  console.log("BUCKET ->", bucketName);
  console.log("TABLE -> ", tableName);

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

  return new PriceService(s3, bucketName, db, tableName);
}

export interface IPriceService {
  getPrices: (date: Date, fuelType: FuelType) => Promise<DayPrices | null>;
  updatePriceCache: (fueltype: FuelType) => Promise<void>;
  fetchPrices: (fuelType: FuelType) => Promise<void>;
}
export class PriceService implements IPriceService {
  private readonly s3: S3;
  private readonly s3Bucket: string;

  private readonly db: DynamoDB.DocumentClient;
  private readonly tableName: string;

  constructor(
    s3: S3,
    s3Bucket: string,
    db: DynamoDB.DocumentClient,
    tableName: string
  ) {
    this.s3 = s3;
    this.s3Bucket = s3Bucket;
    this.db = db;
    this.tableName = tableName;
  }
  /**
   * Get price for the date requested, the day before, and the day after if possible
   */
  async getPrices(date: Date, fuelType: FuelType): Promise<DayPrices | null> {
    const dayDiff = Math.abs(dateDiffInDays(new Date(), date));
    const fuelprices: OkPrices = { historik: [] };
    if (dayDiff < 7) {
      try {
        console.log("getting data from dynamodb");
        const result = await this.db
          .query({
            TableName: this.tableName,
            KeyConditionExpression: "PK = :PK",
            ExpressionAttributeValues: {
              ":PK": `FUELTYPE#${fuelType}`,
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
    } else {
      console.log("getting data from s3");
      try {
        const jsonData = (
          await this.s3
            .getObject({
              Bucket: this.s3Bucket,
              Key: `prices/${fuelType}`,
            })
            .promise()
        )?.Body?.toString("utf-8");
        const fuelpricesObj = JSON.parse(jsonData || "") as OkPrices;
        if (isArray(fuelpricesObj?.historik)) {
          fuelpricesObj.historik.forEach((item) => {
            fuelprices.historik.push(item);
          });
        }
      } catch (error) {
        console.log("error getting s3 data", error);
      }
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

    const yesterdayDate = new Date();
    const dayOffset = 24 * 60 * 60 * 1000 * 1; // 1 day
    yesterdayDate.setTime(date.getTime() - dayOffset);
    const yesterdayOkPrice = findPrice(fuelprices.historik, yesterdayDate);
    if (yesterdayOkPrice) {
      prices.yesterday = {
        date: yesterdayOkPrice.dato,
        price: yesterdayOkPrice.pris,
      };
    }

    const tomorrowDate = new Date();
    tomorrowDate.setTime(date.getTime() + dayOffset);
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

    const recentHistorik = (priceResp.historik ?? []).slice(-10);
    const recentPrices: OkPrices = {
      historik: recentHistorik,
    };
    const ddbItems = recentPrices.historik.map((price) => {
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
    console.log("ddbitems", ddbItems);

    const resp = await this.db
      .batchWrite({
        RequestItems: {
          [this.tableName]: ddbItems,
        },
      })
      .promise();
    console.log(`${fueltype} DONE`);
  }

  async fetchPrices(fuelType: FuelType): Promise<void> {
    console.log("starting fetch");
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
    console.log("OK status:", resp.status);
    const priceResp = (await resp.json()) as OkPrices;
    priceResp.historik.forEach((price) => {
      delete price.prisExclAfgifterExclMoms;
      delete price.prisExclAfgifterInclMoms;
      delete price.prisExclMoms;
      delete price.varenr;
    });
    console.log(this.s3Bucket);
    await this.s3
      .upload({
        Bucket: this.s3Bucket.toLowerCase(),
        Key: `prices/${fuelType}`,
        Body: JSON.stringify(priceResp),
      })
      .promise();

    const recentHistorik = (priceResp.historik ?? []).slice(-10);
    const recentPrices: OkPrices = {
      historik: recentHistorik,
    };
  }
}
