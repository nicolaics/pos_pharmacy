export type Middleware = (
  context: RequestContext,
  next: () => Promise<Response>
) => Promise<Response>;

export interface RequestContext {
  url: string;
  options: RequestInit;
}

export const AuthMiddleware: Middleware = async (context, next) => {
  const token = sessionStorage.getItem("token");
  console.log("auth token: ", token);

  if (token) {
    context.options.headers = {
      ...context.options.headers,
      Authorization: `Bearer ${token}`,
    };
  }
  
  return await next(); // Proceed to the next middleware
};

export const ApplyMiddleware = async (
  middleware: Middleware,
  context: RequestContext
): Promise<Response> => {
  // Directly call the middleware with the fetch as the "next" function
  return await middleware(context, () => fetch(context.url, context.options));
};