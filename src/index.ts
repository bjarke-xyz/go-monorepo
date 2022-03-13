import { handleGetRequest } from './handlers/getHandler'
import {
  handlePostRequest,
  handleScheduledEvent,
} from './handlers/scheduleHandler'
import { PriceGetter } from './lib/prices'

const priceGetter = new PriceGetter()

addEventListener('fetch', async (event) => {
  if (event.request.method == 'GET') {
    event.respondWith(handleGetRequest(event, event.request, priceGetter))
  } else if (event.request.method == 'POST') {
    event.respondWith(handlePostRequest(event, priceGetter))
  } else {
    event.respondWith(
      new Response('Not found', {
        status: 404,
      }),
    )
  }
})

addEventListener('scheduled', async (event) => {
  event.waitUntil(handleScheduledEvent(event, priceGetter))
})
