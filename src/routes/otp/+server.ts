let otp: string | null = null

// SSE only supports GET request
export async function GET({ url }) {
  const stream = new ReadableStream({
      async start(controller) {
          // You can enqueue multiple data asynchronously here.
          while (!otp) {
            await new Promise(resolve => setTimeout(resolve, 1000))
          }

          controller.enqueue(`data: ${otp}\n\n`)
          controller.close() 
          otp = null
      },
      cancel() {
          // cancel your resources here
      }
  });

  return new Response(stream, {
      headers: {
          // Denotes the response as SSE
          'Content-Type': 'text/event-stream', 
          // Optional. Request the GET request not to be cached.
          'Cache-Control': 'no-cache', 
      }
  })
}

export async function POST({ request }) {
  console.log('request', request)
  const { otp: otpBody } = await request.json()
  otp = otpBody
  return new Response('OK', { status: 200 })
}