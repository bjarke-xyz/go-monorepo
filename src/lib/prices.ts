import { isEqual } from 'lodash'
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
    prisExclAfgifterExclMoms?: number
    prisExclAfgifterInclMoms?: number
    prisExclMoms?: number
    varenr?: number
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

// https://stackoverflow.com/a/15289883
function dateDiffInDays(a: Date, b: Date) {
  const _MS_PER_DAY = 1000 * 60 * 60 * 24
  const utc1 = Date.UTC(a.getFullYear(), a.getMonth(), a.getDate())
  const utc2 = Date.UTC(b.getFullYear(), b.getMonth(), b.getDate())

  return Math.floor((utc2 - utc1) / _MS_PER_DAY)
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
    const dayDiff = Math.abs(dateDiffInDays(new Date(), date))
    let fuelpricesStr: string | null = null
    if (dayDiff < 7) {
      fuelpricesStr = await FUELPRICES.get(`prices:${fuelType}:recent`)
    } else {
      fuelpricesStr = await FUELPRICES.get(`prices:${fuelType}`)
    }

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
      const price = historik.find((price) => {
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
    const priceResp = (await resp.json()) as OkPrices
    priceResp.historik.forEach((price) => {
      delete price.prisExclAfgifterExclMoms
      delete price.prisExclAfgifterInclMoms
      delete price.prisExclMoms
      delete price.varenr
    })
    await FUELPRICES.put(`prices:${fuelType}`, JSON.stringify(priceResp))

    const recentHistorik = (priceResp.historik ?? []).slice(-10)
    const recentPrices: OkPrices = {
      historik: recentHistorik,
    }
    const prevRecentPrices = await FUELPRICES.get(`prices:${fuelType}:recent`)
    await FUELPRICES.put(
      `prices:${fuelType}:recent`,
      JSON.stringify(recentPrices),
    )

    if (prevRecentPrices) {
      try {
        const prevRecentPricesObj = JSON.parse(prevRecentPrices) as OkPrices
        await postToQueue(recentPrices, prevRecentPricesObj, fuelType)
      } catch (error) {
        console.error('error posting to queue:', error)
      }
    }

    return
  }
}

declare const MQ_URL: string
declare const MQ_VHOST: string
declare const MQ_EXCHANGE: string
declare const MQ_USER: string
declare const MQ_PASS: string
async function postToQueue(
  recentPrices: OkPrices,
  prevRecentPrices: OkPrices,
  fuelType: FuelType,
): Promise<void> {
  // if (isEqual(recentPrices, prevRecentPrices)) {
  //   return
  // }
  const mqUrl = `${MQ_URL}/api/exchanges/${MQ_VHOST}/${MQ_EXCHANGE}/publish`
  const body = {
    properties: {},
    routing_key: 'test',
    payload: JSON.stringify({ recentPrices, prevRecentPrices, fuelType }),
    payload_encoding: 'string',
  }
  console.log(mqUrl)
  const resp = await fetch(mqUrl, {
    method: 'POST',
    body: JSON.stringify(body),
    headers: {
      Authorization: `Basic ${btoa(MQ_USER + ':' + MQ_PASS)}`,
    },
  })
  console.log('Post to queue response: ', resp.status)
}
