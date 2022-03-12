import makeServiceWorkerEnv from 'service-worker-mock'
import { handleGetRequest } from '../src/handlers/getHandler'
import { DayPrices, FuelType, IPriceGetter } from '../src/lib/prices'

declare let global: any

class PriceGetterStub implements IPriceGetter {
  getPrices(date: Date, fuelType: FuelType): Promise<DayPrices | null> {
    return new Promise((resolve) => {
      resolve({
        today: {
          date: '2022-03-12T00:00:00',
          price: 14.79,
        },
        yesterday: {
          date: '2022-03-11T00:00:00',
          price: 14.79,
        },
        tomorrow: null,
      })
    })
  }

  refreshCache(): Promise<void> {
    return new Promise((resolve) => resolve())
  }
}

describe('handle', () => {
  beforeEach(() => {
    Object.assign(global, makeServiceWorkerEnv())
    jest.resetModules()
  })

  test('handle GET', async () => {
    const priceGetter = new PriceGetterStub()
    const result = await handleGetRequest(
      new Request('/', { method: 'GET' }),
      priceGetter,
    )
    expect(result.status).toEqual(200)
    const text = (await result.json()) as { message: string }
    expect(text.message).toContain(
      'Today, the price of Unleaded octane 95 is 14.79 kroner.',
    )
  })
})
