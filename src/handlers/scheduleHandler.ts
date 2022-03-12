import { PriceGetter } from '../lib/prices'

export async function handlePostRequest(
  request: Request,
  priceGetter: PriceGetter,
): Promise<Response> {
  const key = 'e3e8528c-075b-4eca-b372-c800646a24b7'
  const authHeader = request.headers.get('authorization')
  if (authHeader !== key) {
    return new Response(null, {
      status: 403,
    })
  }
  await refreshCache(priceGetter)
  return new Response('OK')
}

export async function handleScheduledEvent(
  event: ScheduledEvent,
  priceGetter: PriceGetter,
): Promise<void> {
  await refreshCache(priceGetter)
}

async function refreshCache(priceGetter: PriceGetter): Promise<void> {
  try {
    await priceGetter.refreshCache('Unleaded95')
    // await priceGetter.refreshCache('Octane100')
    // await priceGetter.refreshCache('Diesel')
  } catch (error) {
    console.log('Error during cache refresh:', error)
  }
}
