import { Context, SQSEvent } from "aws-lambda";
import { createPriceService, FuelType, OkPrices } from "../lib/prices";

const priceService = createPriceService();

export async function main(event: SQSEvent, context: Context): Promise<any> {
  console.log("event ->", event);
  console.log("context ->", context);
  const items: {
    fueltype: FuelType;
    priceChunk: OkPrices["historik"];
  }[] = [];
  event.Records.forEach((record) => {
    items.push(JSON.parse(record.body));
  });
  items.forEach(async (item) => {
    await priceService.doCacheWrite(item.fueltype, item.priceChunk);
  });
  return {
    body: "",
    statusCode: 200,
  };
}
