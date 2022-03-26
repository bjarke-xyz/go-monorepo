import { PriceGetter } from '../lib/prices'

export const CACHE_REFRESH_CRON = '0 */1 * * *'

declare const CACHE_REFRESH_KEY: string

export async function requestHandlerCacheRefresh(
  event: FetchEvent,
  priceGetter: PriceGetter,
): Promise<Response> {
  const request = event.request
  const authHeader = request.headers.get('authorization')
  if (authHeader !== CACHE_REFRESH_KEY) {
    return new Response(null, {
      status: 403,
    })
  }
  event.waitUntil(refreshCache(priceGetter))
  return new Response('OK')
}

export async function scheduledHandlerCacheRefresh(
  event: ScheduledEvent,
  priceGetter: PriceGetter,
): Promise<void> {
  event.waitUntil(refreshCache(priceGetter))
}

async function refreshCache(priceGetter: PriceGetter): Promise<void> {
  try {
    await priceGetter.refreshCache('Unleaded95')
    await priceGetter.refreshCache('Octane100')
    await priceGetter.refreshCache('Diesel')
  } catch (error) {
    console.log('Error during cache refresh:', error)
  }
}
