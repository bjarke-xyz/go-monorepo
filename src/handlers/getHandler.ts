import { getText, Language } from '../lib/localization'
import { FuelType, PriceGetter } from '../lib/prices'

export async function handleGetRequest(
  request: Request,
  priceGetter: PriceGetter,
): Promise<Response> {
  const { date, fuelType, language } = parseArguments(request)
  const value = await priceGetter.getPrices(date, fuelType)
  if (!value) {
    return new Response('asdf', {
      status: 404,
    })
  }
  const response = {
    message: getText(value, fuelType, language),
    prices: [value],
  }
  const json = JSON.stringify(response)
  return new Response(json, {
    headers: {
      'Content-Type': 'application/json',
    },
  })
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
