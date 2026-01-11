package humaserverless

type HttpMiddleware = func(next HttpHandlerFunc) HttpHandlerFunc

func NoopHttpMiddleware(
	next HttpHandlerFunc,
) HttpHandlerFunc {
	return next
}

func HttpMiddlewareChain(
	middlewares ...HttpMiddleware,
) HttpMiddleware {
	chain := func(next HttpHandlerFunc) HttpHandlerFunc {
		return next
	}

	for i := len(middlewares) - 1; i >= 0; i-- {
		// Capture to avoid closure issues
		currentChain := chain
		currentMiddleware := middlewares[i]

		chain = func(next HttpHandlerFunc) HttpHandlerFunc {
			return currentMiddleware(currentChain(next))
		}
	}

	return chain
}
