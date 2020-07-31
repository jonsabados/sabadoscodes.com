<template>
  <div class="article">
    <h1>X-Ray + golang http clients</h1>
    <hr />
    <p>
      I mentioned it a bit when I wrote about <router-link :to="{name: 'article', params: {id: 1}}">setting up a support email address</router-link>,
      but <a href="https://aws.amazon.com/xray/" target="_blank">X-Ray</a> is a great way to gain insight into your
      applications performance. For a bunch of the AWS services the go clients can more or less automagically integrate
      with X-Ray, but its great to have stats on other things as well. Since calls to third party services can be
      expensive they definitely fall into the "lets keep stats on this" category. At least for now all the third party
      calls that sabadoscodes.com will be making will be done over rest so some sort of X-Ray integration on http
      clients to capture those stats would be nice. Turns out its actually not to difficult!
    </p>
    <p>
      Getting X-Ray tracking on http calls isn't terribly difficult by itself. First you need to get an http client
      that is X-Ray aware, which is not hard at all - Amazon provides functionality to wrap a client in their
      <a href="https://github.com/aws/aws-xray-sdk-go" target="_blank">xray sdk</a>. Simply pass whatever http client
      you are using to <code>xray.Client</code> and that is taken care of. But, that isn't enough - you can make http
      calls with that client all day and night but until the request is associated with a context that has had the
      required x-ray love it won't report. So, every request you make needs to either be created via
      <code>http.NewRequestWithContext</code> or come from calling <code>WithContext</code> on a context-less request.
      Not hard to do, but it is tedious and you have to actually remember to do, and that is the type of thing I am
      terrible at. I also don't want all of the various places I use an http client to have to be aware of the fact
      that the client should be x-ray aware, so how can this be made a bit more automagic?
      <a href="https://github.com/jonsabados/sabadoscodes.com/blob/auth_works/backend/src/go/httputil/httputil.go" target="_blank">This</a> is my solution:
    </p>
    <pre class="code-block">
package httputil

import (
  "context"
  "github.com/aws/aws-xray-sdk-go/xray"
  "net/http"
)

type HTTPClientFactory func(ctx context.Context) *http.Client

type contextAppendingTransport struct {
  ctx     context.Context
  wrapped http.RoundTripper
}

func (c *contextAppendingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
  withCtx := r.WithContext(c.ctx)
  return c.wrapped.RoundTrip(withCtx)
}

func NewContextAppendingTransport(ctx context.Context, toWrap http.RoundTripper) http.RoundTripper {
  if toWrap == nil {
    toWrap = http.DefaultTransport
  }
  return &contextAppendingTransport{ctx, toWrap}
}

func NewXRAYAwareHTTPClientFactory(baseClient *http.Client) HTTPClientFactory {
  return func(ctx context.Context) *http.Client {
    xrayClient := xray.Client(baseClient)
    xrayClient.Transport = NewContextAppendingTransport(ctx, xrayClient.Transport)
    return xrayClient
  }
}

func DefaultHttpClient(_ context.Context) *http.Client {
  return http.DefaultClient
}
    </pre>
    <p>
      The first thing in here is defining a type, <code>HTTPClientFactory</code>. I'll pass one of those as a dependency
      anywhere that needs access to an HTTPClient and use that rather than manually creating things - why should
      things making http calls have to be aware that they are running in an enviornment where X-Ray is in use?
      This also has benefits when it come to unit testing - while the XRay aware http client doesn't blow up if its
      running in an env without X-Ray setup it does complain loudly which can be annoying, so there is a
      <code>DefaultHttpClient</code> function that can be used as a factory when just the default client is needed.
    </p>
    <p>
      Next comes <code>contextAppendingTransport</code> and a function to retrieve one, NewContextAppendingTransport.
      The idea here is to wrapping an existing transport with a thing that will ensure requests are always associated
      with a context. While this helps to make X-Ray happy it is also useful on its own, since there are a couple of
      reasons to associate a request with a context - apart from X-Ray fun context cancellation is respected which
      is suuuuuuper useful in lambda land. All lambda's have an execution timeouts so you need to make sure you have
      some sort of timeout on all of your http calls, which can get tricky if there are multiple calls. But the context
      that comes into the lambda will actually have a deadline, so if you associate your http calls to said context
      you don't have to worry about manually setting timeouts.
    </p>
    <p>
      And then comes <code>NewXRAYAwareHTTPClientFactory</code>. There really isn't too much to it, when it returns a
      client it simply wraps the base client with an xray client and then wraps its transport with a context appending
      transport. Getting the order of that correct took me a little bit for some reason, I started out by wrapping the
      transport on the base client and then wrapping that with the X-Ray client but that resulted in the context being
      appended to the request after the X-Ray clients transport had done its thing. Now, so long as I always get
      http clients via the factory I can't forget to append a context (its required by the factory), and http calls will
      automagically show up in X-Ray, and will looks something like:
    </p>
    <img src="./xray-with-http-call.png" alt="x-ray trace screenshot" class="container-fluid"/>
  </div>
</template>
