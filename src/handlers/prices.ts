import { format, parse, parseISO, subDays } from "date-fns";
import { getErrorText, getText, Language } from "../lib/localization";
import { FuelType } from "../lib/models";
import { PriceRepository } from "../lib/price-repository";
import { IttyRequest } from "../types";

export async function getAllPrices(
  request: IttyRequest,
  context: EventContext<any, any, any>,
  priceRepository: PriceRepository
): Promise<Response> {
  const cacheUrl = new URL(request.url);
  const cacheKey = new Request(cacheUrl.toString(), request);
  const cache = caches.default;

  let response = await cache.match(cacheKey);
  if (!response) {
    const formatStr = "yyyy-MM-dd";
    const now = new Date();
    const fromStr =
      request.query?.["from"] ?? format(subDays(now, 365), formatStr);
    const toStr = request.query?.["to"] ?? format(now, formatStr);
    const fuelType = parseFuelType(request.query?.["type"]);
    const refDate = new Date();
    const from = parse(fromStr, formatStr, refDate);
    const to = parse(toStr, formatStr, refDate);
    const prices = await priceRepository.getAllPrices(fuelType, from, to);

    response = new Response(JSON.stringify(prices), {
      headers: {
        "Content-Type": "application/json",
      },
    });

    response.headers.append("Cache-Control", "s-maxage=900");
    response.headers.append("Access-Control-Allow-Origin", "*");
    context.waitUntil(cache.put(cacheKey, response.clone()));
  }
  return response;
}

export async function getPrices(
  request: IttyRequest,
  context: EventContext<any, any, any>,
  priceRepository: PriceRepository
): Promise<Response> {
  const cacheUrl = new URL(request.url);
  const cacheKey = new Request(cacheUrl.toString(), request);
  const cache = caches.default;

  let response = await cache.match(cacheKey);
  if (!response) {
    const { date, fuelType, language } = parseArguments(request);
    const prices = await priceRepository.getPrices(fuelType, date);
    if (prices) {
      const responseObj = {
        message: getText(prices, fuelType, language),
        prices,
      };

      response = new Response(JSON.stringify(responseObj), {
        headers: {
          "Content-Type": "application/json",
        },
      });
    } else {
      response = new Response(
        JSON.stringify({
          message: getErrorText(language),
        }),
        {
          status: 404,
        }
      );
    }

    response.headers.append("Cache-Control", "s-maxage=900");
    response.headers.append("Access-Control-Allow-Origin", "*");
    context.waitUntil(cache.put(cacheKey, response.clone()));
  }
  return response;
}

function parseArguments(request: IttyRequest): {
  date: Date;
  fuelType: FuelType;
  language: Language;
  noCache: boolean;
} {
  const date = parseDate(request.query?.["now"]);
  const fuelType = parseFuelType(
    request.query?.["type"] ?? request.query?.["fueltype"]
  );
  const noCache = parseNoCache(
    request.query?.["nocache"] ?? request.query?.["noCache"]
  );
  const language = parseLanguage(request.query?.["lang"]);

  return { date, fuelType, language, noCache };
}

function parseDate(dateStr?: string): Date {
  let date = new Date();
  if (dateStr) {
    const parsedDate = parse(dateStr, "yyyy-MM-dd", date);
    if (!isNaN(parsedDate as any)) {
      date = parsedDate;
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

export async function fetchData(
  request: IttyRequest | ScheduledEvent,
  context: EventContext<any, any, any>,
  priceRepository: PriceRepository
): Promise<Response> {
  const promises = [
    priceRepository.fetchAndStoreJsonData("Unleaded95"),
    priceRepository.fetchAndStoreJsonData("Diesel"),
    priceRepository.fetchAndStoreJsonData("Octane100"),
  ];
  context.waitUntil(Promise.all(promises));
  return new Response("OK");
}

export async function updateData(
  request: IttyRequest | ScheduledEvent,
  context: EventContext<any, any, any>,
  priceRepository: PriceRepository
): Promise<Response> {
  const promises = [
    priceRepository.updateKv("Unleaded95"),
    priceRepository.updateKv("Diesel"),
    priceRepository.updateKv("Octane100"),
  ];
  context.waitUntil(Promise.all(promises));
  return new Response("OK");
}
