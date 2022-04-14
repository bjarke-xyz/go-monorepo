import { Callback, Context, S3Event } from "aws-lambda";
import { createPriceService, FuelType } from "../lib/prices";

const priceService = createPriceService();

export async function main(
  event: S3Event,
  context: Context,
  callback: Callback
): Promise<any> {
  console.log("event ->", event);
  console.log("context ->", context);
  const keys = event.Records.map((x) => x.s3.object.key);
  for (const key of keys) {
    console.log("KEY ->", key);
    const fueltype = key.split("/")[1] as FuelType;
    try {
      await priceService.updatePriceCache(fueltype);
    } catch (error) {
      console.log("error updating cache for fuel type", fueltype, error);
    }
  }
  callback(null);
}
