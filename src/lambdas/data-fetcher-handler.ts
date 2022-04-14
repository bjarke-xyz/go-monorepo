import { Callback, Context, EventBridgeEvent } from "aws-lambda";
import { createPriceService } from "../lib/prices";

const priceService = createPriceService();

export async function main(
  event: EventBridgeEvent<any, any>,
  context: Context,
  callback: Callback
): Promise<any> {
  console.log("event ->", event);
  console.log("context ->", context);
  try {
    await priceService.fetchPrices("Unleaded95");
    await priceService.fetchPrices("Octane100");
    await priceService.fetchPrices("Diesel");
  } catch (error) {
    callback(error as any);
    return;
  }
  callback(null);
}
