import { FuelType } from './prices'

declare const FUELPRICES: KVNamespace

export interface Notification {
  fuelType: FuelType
  target: 'discord'
}

export interface DiscordNotification extends Notification {
  url: string
}

export function isDiscordNotification(
  notification: Notification,
): notification is DiscordNotification {
  return notification.target === 'discord'
}

export interface INotificationManager {
  getNotifications: () => Promise<Notification[]>
  sendDiscordNotification: (
    message: string,
    notification: DiscordNotification,
  ) => Promise<void>
}

export class NotificationManager implements INotificationManager {
  async getNotifications(): Promise<Notification[]> {
    const str = await FUELPRICES.get('notifications')
    if (str) {
      return JSON.parse(str) as Notification[]
    } else {
      return []
    }
  }

  async sendDiscordNotification(
    message: string,
    notification: DiscordNotification,
  ): Promise<void> {
    await fetch(notification.url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        username: 'Fuelprices',
        content: message,
      }),
    })
  }
}
