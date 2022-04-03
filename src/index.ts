import { Router } from 'itty-router'
import { handleGetPrices } from './handlers/getPricesHandler'
import {
  CACHE_REFRESH_CRON,
  requestHandlerCacheRefresh,
  scheduledHandlerCacheRefresh,
} from './handlers/cacheRefreshHandler'
import { PriceGetter } from './lib/prices'

const priceGetter = new PriceGetter()

const router = Router()

router.get('/', (request, event) => {
  return handleGetPrices(event, request as Request, priceGetter)
})

router.post('/scheduled/cache', (request, event) => {
  return requestHandlerCacheRefresh(event, priceGetter)
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
