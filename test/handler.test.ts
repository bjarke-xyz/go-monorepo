import makeServiceWorkerEnv from 'service-worker-mock'
import { handleGetRequest } from '../src/handlers/getHandler'
import { DayPrices, FuelType, IPriceGetter } from '../src/lib/prices'

function createPriceGetter(
  todayPrice: number,
  yesterdayPrice?: number,
  tomorrowPrice?: number,
): IPriceGetter {
  return {
    getPrices: function (date: Date, fuelType: FuelType) {
      return new Promise((resolve) =>
        resolve({
          today: {
            date: '2022-03-12T00:00:00',
            price: todayPrice,
          },
          yesterday: {
            date: '2022-03-11T00:00:00',
            price: yesterdayPrice ?? todayPrice,
          },
          tomorrow: {
            date: '2022-03-13T00:00:00',
            price: tomorrowPrice ?? todayPrice,
          },
        }),
      )
    },
    refreshCache: function () {
      return new Promise((resolve) => resolve())
    },
  }
}

function createNoPriceGetter(): IPriceGetter {
  return {
    getPrices: function (date: Date, fuelType: FuelType) {
      return new Promise((resolve) => resolve(null))
    },
    refreshCache: function () {
      return new Promise((resolve) => resolve())
    },
  }
}

declare let global: any
describe('handle', () => {
  beforeEach(() => {
    Object.assign(global, makeServiceWorkerEnv())
    jest.resetModules()
  })

  test('Same prices', async () => {
    const priceGetter = createPriceGetter(14.79)
    const result = await handleGetRequest(
      new Request('/', { method: 'GET' }),
      priceGetter,
    )
    expect(result.status).toEqual(200)
    const text = (await result.json()) as { message: string }
    expect(text.message).toContain('Yesterday the price was the same')
    expect(text.message).toContain('Tomorrow the price will be the same')
  })

  test('Cheaper yesterday prices', async () => {
    const priceGetter = createPriceGetter(14.79, 10)
    const result = await handleGetRequest(
      new Request('/', { method: 'GET' }),
      priceGetter,
    )
    expect(result.status).toEqual(200)
    const text = (await result.json()) as { message: string }
    expect(text.message).toContain('Yesterday the price was lower')
  })

  test('More expensive yesterday prices', async () => {
    const priceGetter = createPriceGetter(14.79, 15)
    const result = await handleGetRequest(
      new Request('/', { method: 'GET' }),
      priceGetter,
    )
    expect(result.status).toEqual(200)
    const text = (await result.json()) as { message: string }
    expect(text.message).toContain('Yesterday the price was higher')
  })

  test('No data found', async () => {
    const priceGetter = createNoPriceGetter()
    const result = await handleGetRequest(
      new Request('/', { method: 'GET' }),
      priceGetter,
    )
    expect(result.status).toEqual(404)
    const text = (await result.json()) as { message: string }
    expect(text.message).toContain('No prices were found')
  })
})
