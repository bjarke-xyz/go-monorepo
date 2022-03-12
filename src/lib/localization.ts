import { DayPrices, FuelType, Price } from './prices'

export type Language = 'da' | 'en'

export function getErrorText(language: Language): string {
  switch (language) {
    case 'da':
      return 'Der blev ikke fundet nogen priser for den dato'
    case 'en':
      return 'No prices were found for that date'
    default:
      return 'No prices were found for that date'
  }
}

export function getText(
  price: DayPrices,
  fuelType: FuelType,
  language: Language,
): string {
  switch (language) {
    case 'da':
      return getTextDanish(price, fuelType)
    case 'en':
      return getTextEnglish(price, fuelType)
    default:
      return getTextEnglish(price, fuelType)
  }
}

function getTextDanish(price: DayPrices, fuelType: FuelType): string {
  const { kroner, orer } = priceToKronerAndOrer(price.today)
  let text = `${fuelTypeToString(
    fuelType,
    'da',
  )} koster ${kroner} kroner og ${orer} ører i dag.`

  if (price.yesterday) {
    const { kroner, orer } = priceToKronerAndOrer(price.yesterday)
    let diffTxt = 'den samme'
    if (price.yesterday.price > price.today.price) {
      diffTxt = 'højere'
    } else if (price.yesterday.price < price.today.price) {
      diffTxt = 'lavere'
    }
    text = `${text} I går var prisen ${diffTxt}: ${kroner} kroner og ${orer} ører.`
  }
  if (price.tomorrow) {
    const { kroner, orer } = priceToKronerAndOrer(price.tomorrow)
    let diffTxt = 'den samme'
    if (price.tomorrow.price > price.today.price) {
      diffTxt = 'højere'
    } else if (price.tomorrow.price < price.today.price) {
      diffTxt = 'lavere'
    }
    text = `${text} I morgen vil prisen være ${diffTxt}: ${kroner} kroner og ${orer} ører.`
  }
  return text
}

function getTextEnglish(price: DayPrices, fuelType: FuelType): string {
  let text = `Today, the price of ${fuelTypeToString(fuelType, 'en')} is ${
    price.today.price
  } kroner.`
  if (price.yesterday) {
    let diffTxt = 'the same'
    if (price.yesterday.price > price.today.price) {
      diffTxt = 'higher'
    } else if (price.yesterday.price < price.today.price) {
      diffTxt = 'lower'
    }
    text = `${text} Yesterday the price was ${diffTxt}: ${price.yesterday.price} kroner.`
  }
  if (price.tomorrow) {
    let diffTxt = 'the same'
    if (price.tomorrow.price > price.today.price) {
      diffTxt = 'higher'
    } else if (price.tomorrow.price < price.today.price) {
      diffTxt = 'lower'
    }
    text = `${text} Tomorrow the price will be ${diffTxt}: ${price.tomorrow.price} kroner.`
  }
  return text
}

function priceToKronerAndOrer(price: Price): { kroner: string; orer: string } {
  const parts = price.price.toString().split('.')
  return {
    kroner: parts[0],
    orer: parts[1],
  }
}

function fuelTypeToString(fuelType: FuelType, language: Language): string {
  switch (language) {
    case 'da': {
      switch (fuelType) {
        case 'Diesel':
          return 'Diesel'
        case 'Octane100':
          return 'Oktan 100'
        case 'Unleaded95':
          return 'Blyfri oktan 95'
        default:
          return 'Blyfri oktan 95'
      }
    }
    case 'en': {
      switch (fuelType) {
        case 'Diesel':
          return 'Diesel'
        case 'Octane100':
          return 'Octane 100'
        case 'Unleaded95':
          return 'Unleaded octane 95'
        default:
          return 'Unleaded octane 95'
      }
    }
  }
}
