import { APIGatewayProxyEventV2, APIGatewayProxyResultV2 } from "aws-lambda";
import { format } from "date-fns";
import { Cache } from "../lib/cache";
import { getErrorText, getText, Language } from "../lib/localization";
import { createPriceService, FuelType } from "../lib/prices";

const cache = new Cache<{ body: string; statusCode: number }>();

export async function main(
  event: APIGatewayProxyEventV2
): Promise<APIGatewayProxyResultV2> {
  console.log("event -> ", event);
  const priceService = createPriceService();

  const { date, fuelType, language, noCache } = parseArguments(event);

  const cacheKey = `${format(date, "yyyy-MM-dd")}:${fuelType}:${language}`;
  const cached = cache.get(cacheKey);
  if (!noCache && cached) {
    console.log("cache hit");
    return {
      ...cached,
      headers: {
        "x-b.xyz-cache": "hit",
      },
    };
  }

  try {
    const value = await priceService.getPrices(date, fuelType);
    if (!value) {
      const val = {
        body: JSON.stringify({
          message: getErrorText(language),
        }),
        statusCode: 404,
      };
      cache.insert(cacheKey, val);
      return {
        ...val,
        headers: {
          "x-b.xyz-cache": "miss",
        },
      };
    }
    const responseObj = {
      message: getText(value, fuelType, language),
      prices: [value],
    };
    const val = {
      body: JSON.stringify(responseObj),
      statusCode: 200,
    };
    cache.insert(cacheKey, val);
    return {
      ...val,
      headers: {
        "x-b.xyz-cache": "miss",
      },
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
  noCache: boolean;
} {
  const date = parseDate(event.queryStringParameters?.["now"]);
  const fuelType = parseFuelType(
    event.queryStringParameters?.["type"] ??
      event.queryStringParameters?.["fueltype"]
  );
  const language = parseLanguage(event.queryStringParameters?.["lang"]);
  const noCache = parseNoCache(
    event.queryStringParameters?.["nocache"] ??
      event.queryStringParameters?.["noCache"]
  );

  return {
    date,
    fuelType,
    language,
    noCache,
  };
}

function parseNoCache(noCacheStr: string | null | undefined): boolean {
  let noCache = false;
  if (noCacheStr) {
    if (noCacheStr.toLowerCase() == "false") {
      noCache = false;
    } else {
      noCache = true;
    }
  }
  return noCache;
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
