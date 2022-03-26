import {
  INotificationManager,
  isDiscordNotification,
} from '../lib/notifications'
import { FuelType, OkPrices } from '../lib/prices'

export async function priceChangeEventHandler(
  event: FetchEvent | null,
  request: Request,
  notificationManager: INotificationManager,
  eventHandlerKey: string,
): Promise<Response> {
  const auth = request.headers.get('Authorization')
  if (!auth || auth !== eventHandlerKey) {
    return new Response('', { status: 401 })
  }
  const payload = (await request.json()) as {
    recentPrices: OkPrices
    prevRecentPrices: OkPrices
    fuelType: FuelType
  }
  const now = new Date()
  const nowDateStr = `${now.getFullYear()}-${now
    .getMonth()
    .toString()
    .padStart(2, '0')}-${now.getDate().toString().padStart(2, '0')}T00:00:00`

  const notifications = await notificationManager.getNotifications()
  let notificationsSent = 0

  const todayPrice = payload.recentPrices.historik.find(
    (x) => x.dato === nowDateStr,
  )
  const prevTodayPrice = payload.prevRecentPrices.historik.find(
    (x) => x.dato === nowDateStr,
  )
  if (!todayPrice || !prevTodayPrice) {
    return new Response(JSON.stringify({ notificationsSent }), { status: 200 })
  } else if (todayPrice.pris === prevTodayPrice.pris) {
    return new Response(JSON.stringify({ notificationsSent }), { status: 200 })
  }

  for (const notification of notifications) {
    if (notification.fuelType === payload.fuelType) {
      notificationsSent++
      switch (notification.target) {
        case 'discord':
          if (isDiscordNotification(notification)) {
            await notificationManager.sendDiscordNotification(
              `Price of ${payload.fuelType} for ${nowDateStr} has changed! Previous: ${prevTodayPrice.pris}, current: ${todayPrice.pris}`,
              notification,
            )
          }
      }
    }
  }

  return new Response(JSON.stringify({ notificationsSent }), { status: 200 })
}
