import { APIGatewayProxyEventV2, APIGatewayProxyResultV2 } from "aws-lambda";
import { getErrorText, getText, Language } from "../lib/localization";
import { createPriceService, FuelType } from "../lib/prices";

export async function main(
  event: APIGatewayProxyEventV2
): Promise<APIGatewayProxyResultV2> {
  console.log("event -> ", event);
  const priceService = createPriceService();

  const { date, fuelType, language } = parseArguments(event);

  try {
    const value = await priceService.getPrices(date, fuelType);
    if (!value) {
      return {
        body: JSON.stringify({
          message: getErrorText(language),
        }),
        statusCode: 404,
      };
    }
    const responseObj = {
      message: getText(value, fuelType, language),
      prices: [value],
    };
    return {
      body: JSON.stringify(responseObj),
      statusCode: 200,
    };
  } catch (error) {
    console.log("error getting prices", error);
    return {
      body: JSON.stringify({
        message: getErrorText(language),
      }),
      statusCode: 500,
    };
  }
}

function parseArguments(event: APIGatewayProxyEventV2): {
  date: Date;
  fuelType: FuelType;
  language: Language;
} {
  const date = parseDate(event.queryStringParameters?.["now"]);
  const fuelType = parseFuelType(
    event.queryStringParameters?.["type"] ??
      event.queryStringParameters?.["fueltype"]
  );
  const language = parseLanguage(event.queryStringParameters?.["lang"]);

  return {
    date,
    fuelType,
    language,
  };
}

function parseDate(dateStr: string | null | undefined): Date {
  const date = new Date();
  if (dateStr) {
    const parsedDate = Date.parse(dateStr);
    if (!isNaN(parsedDate)) {
      date.setTime(parsedDate);
    }
  }
  return date;
}

function parseLanguage(languageStr: string | null | undefined): Language {
  switch (languageStr?.toLowerCase()) {
    case "da":
      return "da";
    case "en":
      return "en";
    default:
      return "en";
  }
}

function parseFuelType(fuelTypeStr: string | null | undefined): FuelType {
  switch (fuelTypeStr?.toLowerCase()) {
    case "unleaded95":
      return "Unleaded95";
    case "octane100":
      return "Octane100";
    case "diesel":
      return "Diesel";
    default:
      return "Unleaded95";
  }
}
