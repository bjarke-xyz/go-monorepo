import { Callback, Context, EventBridgeEvent } from "aws-lambda";
import { createPriceService, FuelType } from "../lib/prices";

const priceService = createPriceService();

export async function main(
  event: EventBridgeEvent<any, any>,
  context: Context,
  callback: Callback
): Promise<any> {
  console.log("event ->", event);
  console.log("context ->", context);
  let fuelTypesToFetch: FuelType[] = ["Unleaded95", "Octane100", "Diesel"];
  if ((event as any).fueltype) {
    switch (((event as any).fueltype as string).toLowerCase()) {
      case "unleaded95":
        fuelTypesToFetch = ["Unleaded95"];
        break;
      case "octane100":
        fuelTypesToFetch = ["Octane100"];
        break;
      case "diesel":
        fuelTypesToFetch = ["Diesel"];
        break;
    }
  }
  try {
    console.log("fetching", fuelTypesToFetch);
    fuelTypesToFetch.forEach(async (fueltype) => {
      await priceService.fetchPrices(fueltype);
    });
  } catch (error) {
    callback(error as any);
    return;
  }
  callback(null);
}
