import { Router } from 'itty-router'
import { handleGetPrices } from './handlers/getPricesHandler'
import {
  CACHE_REFRESH_CRON,
  requestHandlerCacheRefresh,
  scheduledHandlerCacheRefresh,
} from './handlers/cacheRefreshHandler'
import { PriceGetter } from './lib/prices'
import { priceChangeEventHandler } from './handlers/priceChangeEventHandler'
import { NotificationManager } from './lib/notifications'

const priceGetter = new PriceGetter()

const router = Router()

router.get('/', (request, event) => {
  return handleGetPrices(event, request as Request, priceGetter)
})

router.post('/scheduled/cache', (request, event) => {
  return requestHandlerCacheRefresh(event, priceGetter)
})

const notificationGetter = new NotificationManager()
declare const EVENT_HANDLER_KEY: string
router.post('/events/pricechanges', (request: Request, event) => {
  return priceChangeEventHandler(
    event,
    request,
    notificationGetter,
    EVENT_HANDLER_KEY,
  )
})

router.all('*', () => {
  return new Response('Not found', { status: 404 })
})

addEventListener('fetch', (event) => {
  event.respondWith(router.handle(event.request, event))
})

addEventListener('scheduled', async (event) => {
  switch (event.cron) {
    case CACHE_REFRESH_CRON:
      event.waitUntil(scheduledHandlerCacheRefresh(event, priceGetter))
      break
  }
  console.log('cron processed')
})
