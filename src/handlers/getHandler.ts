import { getErrorText, getText, Language } from '../lib/localization'
import { FuelType, PriceGetter } from '../lib/prices'

export async function handleGetRequest(
  event: FetchEvent,
  priceGetter: PriceGetter,
): Promise<Response> {
  const request = event.request
  const cache = caches.default
  const cacheUrl = new URL(request.url)
  const cacheKey = new Request(cacheUrl.toString(), request)

  let response = await cache.match(cacheKey)

  if (!response) {
    console.log('cache miss')
    const { date, fuelType, language } = parseArguments(request)
    const value = await priceGetter.getPrices(date, fuelType)
    if (!value) {
      const error = {
        message: getErrorText(language),
        prices: [],
      }
      return new Response(JSON.stringify(error), {
        status: 404,
        headers: {
          'Content-Type': 'application/json',
        },
      })
    }
    const responseObj = {
      message: getText(value, fuelType, language),
      prices: [value],
    }
    const json = JSON.stringify(responseObj)
    response = new Response(json, {
      headers: {
        'Content-Type': 'application/json',
        'Cache-Control': 's-maxage=60',
      },
    })
    event.waitUntil(cache.put(cacheKey, response.clone()))
  }
  return response
}

function parseArguments(request: Request): {
  date: Date
  fuelType: FuelType
  language: Language
} {
  const date = new Date()
  const url = new URL(request.url)
  const nowStr = url.searchParams.get('now')
  if (nowStr) {
    const parsedDate = Date.parse(nowStr)
    if (!isNaN(parsedDate)) {
      date.setTime(parsedDate)
    }
  }

  const fuelType = parseFuelType(url.searchParams.get('fuelType'))

  const language = parseLanguage(url.searchParams.get('lang'))

  return {
    date,
    fuelType,
    language,
  }
}

function parseLanguage(languageStr: string | null): Language {
  switch (languageStr?.toLowerCase()) {
    case 'da':
      return 'da'
    case 'en':
      return 'en'
    default:
      return 'en'
  }
}

function parseFuelType(fuelTypeStr: string | null): FuelType {
  switch (fuelTypeStr?.toLowerCase()) {
    case 'unleaded95':
      return 'Unleaded95'
    case 'octane100':
      return 'Octane100'
    case 'diesel':
      return 'Diesel'
    default:
      return 'Unleaded95'
  }
}
