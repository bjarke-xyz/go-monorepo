import makeServiceWorkerEnv from 'service-worker-mock'
import { priceChangeEventHandler } from '../src/handlers/priceChangeEventHandler'
import {
  DiscordNotification,
  INotificationManager,
  Notification,
} from '../src/lib/notifications'
import { FuelType, OkPrices } from '../src/lib/prices'

class MockNotificationManager implements INotificationManager {
  getNotifications(): Promise<Notification[]> {
    return new Promise((resolve) => {
      resolve([
        {
          fuelType: 'Unleaded95',
          target: 'discord',
          url: 'blank',
        } as DiscordNotification,
      ])
    })
  }

  sendDiscordNotification(
    message: string,
    notification: DiscordNotification,
  ): Promise<void> {
    return new Promise((resolve) => {
      console.log(`discord notification sent with message ${message}`)
      resolve()
    })
  }
}

function createBody(
  price: number,
  prevPrice: number,
): { recentPrices: OkPrices; prevRecentPrices: OkPrices; fuelType: FuelType } {
  const now = new Date()
  const date = `${now.getFullYear()}-${now
    .getMonth()
    .toString()
    .padStart(2, '0')}-${now.getDate().toString().padStart(2, '0')}T00:00:00`
  const body: {
    recentPrices: OkPrices
    prevRecentPrices: OkPrices
    fuelType: FuelType
  } = {
    recentPrices: {
      historik: [
        {
          dato: date,
          pris: price,
        },
      ],
    },
    prevRecentPrices: {
      historik: [
        {
          dato: date,
          pris: prevPrice,
        },
      ],
    },
    fuelType: 'Unleaded95',
  }
  return body
}

declare let global: unknown
describe('price change event handler', () => {
  beforeEach(() => {
    Object.assign(global, makeServiceWorkerEnv())
    jest.resetModules()
  })

  test('Same price', async () => {
    const body = createBody(10, 10)
    const request = new Request('', {
      headers: {
        Authorization: '1234',
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(body),
    })
    const notificationGetter = new MockNotificationManager()
    const response = await priceChangeEventHandler(
      null,
      request,
      notificationGetter,
      '1234',
    )
    const responseJson = (await response.json()) as {
      notificationsSent: number
    }
    expect(response.status).toEqual(200)
    expect(responseJson.notificationsSent).toEqual(0)
  })

  test('Price raised', async () => {
    const body = createBody(10, 9)
    const request = new Request('', {
      headers: {
        Authorization: '1234',
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(body),
    })
    const notificationGetter = new MockNotificationManager()
    const response = await priceChangeEventHandler(
      null,
      request,
      notificationGetter,
      '1234',
    )
    const responseJson = (await response.json()) as {
      notificationsSent: number
    }
    expect(response.status).toEqual(200)
    expect(responseJson.notificationsSent).toEqual(1)
  })

  test('Price lowered', async () => {
    const body = createBody(9, 10)
    const request = new Request('', {
      headers: {
        Authorization: '1234',
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(body),
    })
    const notificationGetter = new MockNotificationManager()
    const response = await priceChangeEventHandler(
      null,
      request,
      notificationGetter,
      '1234',
    )
    const responseJson = (await response.json()) as {
      notificationsSent: number
    }
    expect(response.status).toEqual(200)
    expect(responseJson.notificationsSent).toEqual(1)
  })
})
