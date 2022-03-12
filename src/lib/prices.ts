declare const FUELPRICES: KVNamespace

export type FuelType = 'Unleaded95' | 'Octane100' | 'Diesel'
function fuelTypeToOkItemNumber(fuelType: FuelType): number {
  switch (fuelType) {
    case 'Unleaded95':
      return 536
    case 'Octane100':
      return 533
    case 'Diesel':
      return 231
    default:
      return 536
  }
}

interface OkPrices {
  historik: {
    dato: string
    pris: number
  }[]
}

export interface Price {
  date: string
  price: number
}

export interface DayPrices {
  today: Price
  yesterday: Price | null
  tomorrow: Price | null
}

export interface IPriceGetter {
  getPrices: (date: Date, fuelType: FuelType) => Promise<DayPrices | null>
  refreshCache: (fuelType: FuelType) => Promise<void>
}
export class PriceGetter implements IPriceGetter {
  /**
   * Get price for the date requested, the day before, and the day after if possible
   */
  async getPrices(date: Date, fuelType: FuelType): Promise<DayPrices | null> {
    const fuelpricesStr = await FUELPRICES.get(`prices:${fuelType}`)
    if (!fuelpricesStr) {
      return null
    }

    const fuelprices = JSON.parse(fuelpricesStr) as OkPrices
    if (!Array.isArray(fuelprices?.historik)) {
      return null
    }

    function findPrice(
      historik: OkPrices['historik'],
      date: Date,
    ): OkPrices['historik'][0] | null {
      const price = fuelprices.historik.find((price) => {
        const [year, month, day] = price.dato
          .split('T')[0]
          .split('-')
          .map((x) => Number(x))
        return (
          year === date.getFullYear() &&
          month === date.getMonth() + 1 &&
          day === date.getDate()
        )
      })
      if (!price) {
        return null
      }
      return price
    }

    const todayOkPrice = findPrice(fuelprices.historik, date)
    if (!todayOkPrice) {
      return null
    }
    const prices: DayPrices = {
      today: {
        date: todayOkPrice.dato,
        price: todayOkPrice.pris,
      },
      tomorrow: null,
      yesterday: null,
    }

    const yesterdayDate = new Date()
    const dayOffset = 24 * 60 * 60 * 1000 * 1 // 1 day
    yesterdayDate.setTime(date.getTime() - dayOffset)
    const yesterdayOkPrice = findPrice(fuelprices.historik, yesterdayDate)
    if (yesterdayOkPrice) {
      prices.yesterday = {
        date: yesterdayOkPrice.dato,
        price: yesterdayOkPrice.pris,
      }
    }

    const tomorrowDate = new Date()
    tomorrowDate.setTime(date.getTime() + dayOffset)
    const tomorrowOkPrice = findPrice(fuelprices.historik, tomorrowDate)
    if (tomorrowOkPrice) {
      prices.tomorrow = {
        date: tomorrowOkPrice.dato,
        price: tomorrowOkPrice.pris,
      }
    }

    return prices
  }

  async refreshCache(fuelType: FuelType): Promise<void> {
    console.log('starting fetch')
    const resp = await fetch(
      'https://www.ok.dk/privat/produkter/benzinkort/prisudvikling/getProduktHistorik',
      {
        method: 'POST',
        body: JSON.stringify({
          varenr: fuelTypeToOkItemNumber(fuelType),
          pumpepris: true,
        }),
        headers: {
          'Content-Type': 'application/json',
        },
      },
    )
    console.log('OK status:', resp.status)
    const priceResp = await resp.text()
    await FUELPRICES.put(`prices:${fuelType}`, priceResp)
    return
  }
}
