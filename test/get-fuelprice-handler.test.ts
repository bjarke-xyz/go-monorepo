import { APIGatewayProxyEventV2 } from "aws-lambda";
import { main } from "../src/lambdas/get-fuelprice-handler";
import * as prices from "../src/lib/prices";
import { DayPrices, FuelType, IPriceService } from "../src/lib/prices";

interface LambdaResponse {
  statusCode: number;
  body: string;
}

class FakePriceService implements IPriceService {
  todayPrice: number;
  yesterdayPrice?: number;
  tomorrowPrice?: number;

  constructor(
    todayPrice: number,
    yesterdayPrice?: number,
    tomorrowPrice?: number
  ) {
    this.todayPrice = todayPrice;
    this.yesterdayPrice = yesterdayPrice;
    this.tomorrowPrice = tomorrowPrice;
  }

  getPrices(date: Date, fuelType: FuelType): Promise<DayPrices | null> {
    return new Promise((resolve) =>
      resolve({
        today: {
          date: "2022-03-12T00:00:00",
          price: this.todayPrice,
        },
        yesterday: {
          date: "2022-03-11T00:00:00",
          price: this.yesterdayPrice ?? this.todayPrice,
        },
        tomorrow: {
          date: "2022-03-13T00:00:00",
          price: this.tomorrowPrice ?? this.todayPrice,
        },
      })
    );
  }

  async updatePriceCache(fueltype: FuelType): Promise<void> {}

  async fetchPrices(fuelType: FuelType): Promise<void> {}
}

class NoDataPriceService implements IPriceService {
  getPrices(date: Date, fuelType: FuelType): Promise<DayPrices | null> {
    return new Promise((resolve) => resolve(null));
  }

  async updatePriceCache(fueltype: FuelType): Promise<void> {}

  async fetchPrices(fuelType: FuelType): Promise<void> {}
}

describe("Get fuelprice handler", () => {
  test("Should return same price response", async () => {
    jest
      .spyOn(prices, "createPriceService")
      .mockReturnValue(new FakePriceService(1));
    const event: APIGatewayProxyEventV2 = {} as any;
    const response = (await main(event)) as LambdaResponse;
    expect(response.statusCode).toEqual(200);
    const body = response.body as string;
    expect(body).toContain("Yesterday the price was the same");
    expect(body).toContain("Tomorrow the price will be the same");
  });

  test("Cheaper yesterday prices", async () => {
    jest
      .spyOn(prices, "createPriceService")
      .mockReturnValue(new FakePriceService(10, 9));
    const event: APIGatewayProxyEventV2 = {} as any;
    const response = (await main(event)) as LambdaResponse;
    expect(response.statusCode).toEqual(200);
    const body = response.body as string;
    expect(body).toContain("Yesterday the price was lower");
  });

  test("More expensive yesterday prices", async () => {
    jest
      .spyOn(prices, "createPriceService")
      .mockReturnValue(new FakePriceService(10, 11));
    const event: APIGatewayProxyEventV2 = {} as any;
    const response = (await main(event)) as LambdaResponse;
    expect(response.statusCode).toEqual(200);
    const body = response.body as string;
    expect(body).toContain("Yesterday the price was higher");
  });

  test("No data found", async () => {
    jest
      .spyOn(prices, "createPriceService")
      .mockReturnValue(new NoDataPriceService());
    const event: APIGatewayProxyEventV2 = {} as any;
    const response = (await main(event)) as LambdaResponse;
    expect(response.statusCode).toEqual(404);
    const body = response.body as string;
    expect(body).toContain("No prices were found");
  });
});
