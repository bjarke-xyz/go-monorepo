import {
  addDays,
  differenceInDays,
  hoursToMilliseconds,
  subDays,
} from "date-fns";
import _, { groupBy, isArray, isNil, reverse, take } from "lodash";
import { WorkerEnv } from "../types";
import {
  DayPrices,
  FuelType,
  fuelTypeToOkItemNumber,
  OkPrices,
} from "./models";

type KvKey = "recent" | "archive";

function getKvKey(fuelType: FuelType, key: KvKey) {
  return `${key}:${fuelType}`;
}

function getR2Key(fuelType: FuelType) {
  return `prices/${fuelType}`;
}

export class PriceRepository {
  constructor(private readonly env: WorkerEnv) {}

  async getPrices(fuelType: FuelType, date: Date): Promise<DayPrices | null> {
    const now = new Date();
    const dateDiff = Math.abs(differenceInDays(date, now));
    let kvKey: KvKey = "recent";
    if (dateDiff > 31) {
      kvKey = "archive";
    }

    const getPriceHelper = async () => {
      return (
        (await this.env.KV_FUELPRICES.get<OkPrices["historik"]>(
          getKvKey(fuelType, kvKey),
          "json"
        )) ?? []
      );
    };

    let fuelprices = await getPriceHelper();
    if (fuelprices.length === 0) {
      console.log("No prices found, updating kv");
      await this.updateKv(fuelType);
      fuelprices = await getPriceHelper();
    }

    const findPrice = (historik: OkPrices["historik"], date: Date) => {
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
    };

    const todayOkPrice = findPrice(fuelprices, date);
    if (!todayOkPrice) {
      return null;
    }
    const prices: DayPrices = {
      today: {
        date: todayOkPrice.dato,
        price: todayOkPrice.pris,
        prevPrices: todayOkPrice.prevPrices ?? [],
      },
      tomorrow: null,
      yesterday: null,
    };

    const yesterdayDate = subDays(date, 1);
    const yesterdayOkPrice = findPrice(fuelprices, yesterdayDate);
    if (yesterdayOkPrice) {
      prices.yesterday = {
        date: yesterdayOkPrice.dato,
        price: yesterdayOkPrice.pris,
        prevPrices: yesterdayOkPrice.prevPrices ?? [],
      };
    }

    const tomorrowDate = addDays(date, 1);
    const tomorrowOkPrice = findPrice(fuelprices, tomorrowDate);
    if (tomorrowOkPrice) {
      prices.tomorrow = {
        date: tomorrowOkPrice.dato,
        price: tomorrowOkPrice.pris,
        prevPrices: tomorrowOkPrice.prevPrices ?? [],
      };
    }

    return prices;
  }

  async updateKv(fuelType: FuelType): Promise<void> {
    const object = await this.env.R2_FUELPRICES.get(getR2Key(fuelType));
    if (object) {
      const okPrices = await object.json<OkPrices>();
      const prices = okPrices["historik"];
      const expirationTtlSeconds = 1 * 60 * 62; // 3720s, 1h and 2 min
      await this.env.KV_FUELPRICES.put(
        getKvKey(fuelType, "archive"),
        JSON.stringify(prices),
        {
          expirationTtl: expirationTtlSeconds,
        }
      );

      const dbRecent = await this.env.KV_FUELPRICES.get<OkPrices["historik"]>(
        getKvKey(fuelType, "recent"),
        "json"
      );

      const recentPrices = take(reverse(prices), 33);
      const recentPricesByDate = groupBy(recentPrices, (x) => x.dato);

      if (dbRecent && isArray(dbRecent)) {
        for (const dbPrice of dbRecent) {
          const recentPrice = recentPricesByDate[dbPrice.dato]?.[0];
          if (recentPrice && recentPrice.pris !== dbPrice.pris) {
            if (!recentPrice.prevPrices) {
              recentPrice.prevPrices = [];
            }
            if (dbPrice.prevPrices && dbPrice.prevPrices.length > 0) {
              recentPrice.prevPrices = [...dbPrice.prevPrices];
            }
            recentPrice.prevPrices.push({
              detectionTimestamp: new Date().toISOString(),
              price: dbPrice.pris,
            });
          }
        }
      }

      await this.env.KV_FUELPRICES.put(
        getKvKey(fuelType, "recent"),
        JSON.stringify(recentPrices),
        {
          expirationTtl: expirationTtlSeconds,
        }
      );
    }
  }

  async fetchAndStoreJsonData(fuelType: FuelType): Promise<void> {
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

    console.log(fuelType, priceResp.historik.length);
    await this.env.R2_FUELPRICES.put(
      getR2Key(fuelType),
      JSON.stringify(priceResp)
    );
  }
}
