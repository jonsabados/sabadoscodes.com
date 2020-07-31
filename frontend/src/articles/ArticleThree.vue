<template>
  <div class="article">
    <h1>User Authentication, V1</h1>
    <hr />
    <p>
      <i>
        I haven't had a ton of time to work on toy project stuff lately so getting user authentication in place took
        some time. Since I wanted to devote as little time and brainpower on auth as possible I went with
        <a href="https://developers.google.com/identity/sign-in/web" target="_blank">Google Sign-In For Websites</a>
        which handles JWT token fetching entirely on the client side, leaving only token verification for the backend
        (yay!). But sometime during the very long stretch of small fits of productivity on this Google updated Chrome
        to block third party cookies by default which
        <a href="https://developers.google.com/identity/sign-in/web/troubleshooting" target="_blank">breaks things</a>,
        hence the V1. At some point I will need to switch to doing a
        <a href="https://developers.google.com/identity/protocols/oauth2/web-server" target="_blank">server side oAuth flow</a>.
      </i>
    </p>
    <hr />
    <h2>What's the plan?</h2>
    <p>
      I already know my plan for exposing back end rest functionality is going to be accomplished by combining
      <a href="https://aws.amazon.com/api-gateway/" target="_blank">API Gateway</a> with
      <a href="https://golang.org/" target="_blank">Go</a> based <a href="https://aws.amazon.com/lambda/" target="_blank">Lambdas</a>,
      so the logical solution would either be to use <a href="https://aws.amazon.com/cognito/" target="_blank">Cognito</a>
      or do a custom authorizer. While it wouldn't be the end of the world to require users to sign up for a site
      specific account it would be more work, and it would also require users to have one more username to remember.
      Even with a username/password management system in place I don't even want to track one more set of credentials,
      so that is out. Instead I want to do something like login with Facebook, or login with Google. Both of which could
      be done with either Cognito or a custom authorizer.
    </p>
    <p>
      Since this is my own personal little project "Because I Feel Like It" is a perfectly fine reason for any technical
      decision, and for whatever reason fussing with Cognito just didn't seem like as much fun as doing a custom
      authorizer so decision made there. While I could theoretically support a number of different services down the
      line I have to choose one to start.
      <a href="https://developers.google.com/identity/sign-in/web" target="_blank">Google Sign-In For Websites</a>
      can be just dropped in on the front end which would leave only token validation on the back end so that'll be
      my source of users.
    </p>
    <hr />
    <h2>UI Bits</h2>
    <p>
      There really isn't anything ground breaking or exciting about the UI side of things, so I'm not going to go into
      depth on the UI. But, most of the UI bits related to auth live inside the
      <a href="https://github.com/jonsabados/sabadoscodes.com/tree/auth_works/frontend/src/user">frontend/src/user</a>
      directory. The components are:
    </p>
    <ul>
      <li><a href="https://github.com/jonsabados/sabadoscodes.com/blob/auth_works/frontend/src/user/google.ts">google.ts</a></li>
      <li><a href="https://github.com/jonsabados/sabadoscodes.com/blob/auth_works/frontend/src/user/UserStore.ts">UserStore.ts</a></li>
      <li><a href="https://github.com/jonsabados/sabadoscodes.com/blob/auth_works/frontend/src/user/SignIn.vue">SignIn.vue</a></li>
      <li><a href="https://github.com/jonsabados/sabadoscodes.com/blob/auth_works/frontend/src/user/self.ts">self.ts</a></li>
    </ul>
    <p>
      Some wheels are fun to re-invent while others, like getting google login working with a vue application really
      are not. So I'm making use of <a href="https://www.npmjs.com/package/vue-google-login" target="_blank">vue-google-login</a>
      to do all the heavy lifting there. Unfortunately it does not offer any type definitions and I live in typescript land so
      <a href="https://github.com/jonsabados/sabadoscodes.com/blob/auth_works/frontend/src/user/google.ts">google.ts</a>
      exists mostly to wrap the non-typescript stuff.
    </p>
    <p>
      Rather than have a single vuex store for all of the front end state I've been breaking things down into specific
      modules. <a href="https://github.com/jonsabados/sabadoscodes.com/blob/auth_works/frontend/src/user/UserStore.ts">UserStore.ts</a>
      is the module for the current user (or non user). Its a little rough right now, but works as as starting point to
      be iterated upon. The one noteworthy thing about it is that it needs to be initialized to hook into the Google web
      sign on lifecycle so hooks were added in
      <a href="https://github.com/jonsabados/sabadoscodes.com/blob/auth_works/frontend/src/App.vue#L38">App.vue</a> to
      do just that.
    </p>
    <p>
      <a href="https://github.com/jonsabados/sabadoscodes.com/blob/auth_works/frontend/src/user/SignIn.vue">SignIn.vue</a>
      is the component responsible for showing the log in/log out action.
      <a href="https://github.com/jonsabados/sabadoscodes.com/blob/auth_works/frontend/src/navigation/MainNav.vue#L15">MainNav.vue</a>
      pulls it into the mix.
    </p>
    <p>
      <a href="https://github.com/jonsabados/sabadoscodes.com/blob/auth_works/frontend/src/user/self.ts">self.ts</a>
      defines types and functions for interacting with the self endpoint. Since its not vue specific no reason not to
      extract it out to its own thing.
    </p>
    <hr />
    <h2>The Backend</h2>
    <p>
      While the UI portion was pretty straight forward the backend was another story & required a decent amount of
      discovery so I'm going to talk about the areas that I thought were cool, or where where I learned things.
    </p>
    <h3>API groundwork - the API gateway</h3>
    <p>
      The first step for getting the backend going was to lay the groundwork for API gateway. This was all done in
      Terraform and can be found in <a href="https://github.com/jonsabados/sabadoscodes.com/blob/auth_works/infrastructure/api.tf">this file</a>.
      This terraform sets up the following items:
    </p>
    <ul>
      <li><a href="https://github.com/jonsabados/sabadoscodes.com/blob/auth_works/infrastructure/api.tf#L4">Creates the API Gateway</a></li>
      <li><a href="https://github.com/jonsabados/sabadoscodes.com/blob/auth_works/infrastructure/api.tf#L1">ACM managed certificate for api.sabadoscodes.com</a></li>
      <li><a href="https://github.com/jonsabados/sabadoscodes.com/blob/auth_works/infrastructure/api.tf#L6">Creates the validation record for the certificate in Route53</a></li>
      <li><a href="https://github.com/jonsabados/sabadoscodes.com/blob/auth_works/infrastructure/api.tf#L18">Creates a custom domain name for api.sabadoscodes.com</a></li>
      <li><a href="https://github.com/jonsabados/sabadoscodes.com/blob/auth_works/infrastructure/api.tf#L43">Creates the api.sabadoscodes.com DNS record in Route53</a></li>
      <li><a href="https://github.com/jonsabados/sabadoscodes.com/blob/auth_works/infrastructure/api.tf#L23">Creates a deployment</a></li>
      <li><a href="https://github.com/jonsabados/sabadoscodes.com/blob/auth_works/infrastructure/api.tf#L37">Creates a stage</a></li>
      <li><a href="https://github.com/jonsabados/sabadoscodes.com/blob/auth_works/infrastructure/api.tf#L51">Maps the custom domain to the stage</a></li>
    </ul>
    <p>
      Sooo... there's a decent amount that goes into creating an API gateway and exposing it to the world. Even though
      the ACM certificate is first in the file the real starting point is the API gateway itself, which isn't much more
      than a name. While you could technically hit API gateway resources using invocation URLs for stages it wouldn't be
      practical or pleasant so the next step is to get a domain name working for it via an
      <code>aws_api_gateway_domain_name</code>, which requires a cert. So the next step is to create an ACM managed
      certificate using <code>aws_acm_certificate</code> and then verify it. The verification is done by adding a DNS record,
      and I'm using Route53, so this can actually be done in terraform which is what the
      <code>aws_route53_record.api_cert_verification_record</code> resource is all about. At this point the gateway
      will theoretically respond to requests based on the custom domain name (in my case <code>api.sabadoscodes.com</code>),
      but the domain name needs a backing DNS entry to resolve. A quick terraform entry for <code>aws_route53_record.api</code>
      accomplishes that by setting a CNAME up pointing to <code>aws_api_gateway_domain_name.api.cloudfront_domain_name</code>.
    </p>
    <p>
      After this comes the need to deal with API Gateway's stage/deployment shenanigans, which is the rest of the linked
      Terraform. For better or worse (I think worse) API Gateway has the idea of stages and deployments - there has to
      be a use case for this, but in sane envs where you have fully separate dev/qa/production environments they are
      just a royal pain in the ass. In my case I was initially having to manually deploy the API every time I made
      any sort of change due to Terraform resetting the stages deployment back to the version stored in the state. I got
      around this by modifying the <resource>aws_api_gateway_deployment.main</resource> to contain a "deployed_at"
      variable and modifying its life cycle to create new resources before destroying old ones. The result of this is
      any time I run terraform a new deployment is created and then that deployment is mapped to the (only) stage which
      is associated with the custom domain name. Whew. I hope I don't have to think about that ever again.
    </p>
    <h3>Parsing JWT tokens - they are just signed base64 encoded json blobs</h3>
    <p>
      Once a user has authenticated on the front end client side code will have a JWT token signed by google available
      to it. This token will be included as a bearer token on requests made to the backend, which will need to verify
      it. Google provides SDKs to do this for
      <a href="https://developers.google.com/identity/sign-in/android/backend-auth" target="_blank">most languages</a>,
      however the only method provided for go involves making a REST call to Googles api to do the verification which is
      pretty lame (I lost track of the exact call to do this, but am pretty confident I didn't just amke it up). There
      is some <a href="https://godoc.org/golang.org/x/oauth2/jws#Verify" target="_blank">canned stuff</a> for verifying
      JWT tokens in the x/oath2 package, but the functionality to parse tokens drops information I care about on the
      floor. Turns out there is code out there to do google id token verification without making calls to google (well
      except for fetching public certs), but since this is my own project and its something that is interesting to me
      I can actually throw out the don't re-invent the wheel rule that I would strictly apply in a professional setting.
    </p>
    <p>
      The reason that I'm so eager to re-invent the wheel and do the JWT token verification myself is because I've dealt
      with JWT tokens for quite a while, but have always let underlying frameworks do the needful. This is a good thing,
      as security stuff is hard to get correct and the costs of messing it up are high. Here's my low risk opportunity
      to do some learning :). So where to start?
    </p>
    <p>
      While I don't know a ton about JWT tokens I do know that they are signed by one parties private keys and can be
      verified using their public keys, and the public keys for Google token signing can be found at
      <a href="https://www.googleapis.com/oauth2/v1/certs" target="_blank">https://www.googleapis.com/oauth2/v1/certs</a>.
      I could potentially just snag them and embed them in the code, but that seems fragile in the event of changes on
      Google's side. I also know that these things aren't going to change often so some sort of caching makes sense.
      And finally I'm going to want to be able to test the JWT verification code when I get to it, and I don't have
      Google's private keys, so whatever I implement needs to be done in a way that I can inject public keys belonging
      to private keys that I generate for tests.
    </p>
    <p>
      First up comes fetching the certs - while I could mix the caching logic into this that just seems like a terrible
      idea, so my <a href="https://github.com/jonsabados/sabadoscodes.com/blob/auth_works/backend/src/go/auth/google.go#L129">interface for this</a>
      is just going to be a function that returns a list of certificates and an expiration time. The
      <a href="https://github.com/jonsabados/sabadoscodes.com/blob/auth_works/backend/src/go/auth/google.go#L138">actual implementation</a>
      will have the URL to fetch from and an http client factory injected as dependencies through a higher order function.
      This is so the implementation can be <a href="https://github.com/jonsabados/sabadoscodes.com/blob/auth_works/backend/src/go/auth/google_test.go#L48">tested</a>
      and so that I don't have to care about things like associating requests to contexts or what http client I am using
      (see <router-link :to="{name: 'article', params: {id: 2}}">this article</router-link> for more there). Otherwise
      cert fetching code is just a basic rest call, so nothing to crazy to talk about there.
    </p>
    <p>
      Next up is actually validating the tokens. In hind site my
      <a href="https://github.com/jonsabados/sabadoscodes.com/blob/auth_works/backend/src/go/auth/google.go#L118">implementation</a>
      does too much - it deals with both cert validation as well as caching. If I'm ever in that code again I will likely
      extract a thing to cache the certs out, but the code works and is tested so leaving as is. So what all needs
      to happen to validate a JWT token? First question to answer is what the hell is a JWT to begin with? The answer
      to that is it is a combination of three things stuffed into a single string separated by <code>.</code>'s. Example
      JWT:
    </p>
    <p class="code-block">
eyJhbGciOiJSUzI1NiIsImtpZCI6ImE0MWEzNTcwYjhlM2FlMWI3MmNhYWJjYWE3YjhkMmRiMjA2NWQ3YzEiLCJ0eXAiOiJKV1QifQ.ewogICJpc3MiOiAiZm9vLmJhci5jb20iLAogICJhenAiOiAid2hhdGV2ZXIiLAogICJhdWQiOiAidGVzdHlNY1Rlc3RlcnNvbiIsCiAgInN1YiI6ICIxMjM0NSIsCiAgImVtYWlsIjogInRlc3RAdGVzdC5jb20iLAogICJlbWFpbF92ZXJpZmllZCI6IHRydWUsCiAgImF0X2hhc2giOiAiOVNyS0g1R1J0WHVsMXlScFdDeUxvdyIsCiAgIm5hbWUiOiAiQm9iIE1jVGVzdGVyIiwKICAicGljdHVyZSI6ICJodHRwczovL3NvbWUuZ29vZ2xlLmxpbmsvYmxhaCIsCiAgImdpdmVuX25hbWUiOiAiQm9iIiwKICAiZmFtaWx5X25hbWUiOiAiTWNUZXN0ZXIiLAogICJsb2NhbGUiOiAiZW4iLAogICJpYXQiOiAxNTk1MjgxOTU3LAogICJleHAiOiAxNTk1Mjg1NTU3LAogICJqdGkiOiAiMTIzYWJjNTYiCn0.Fldxl1a0djgcAI3JS2YOlKVhXex50k8Fbbvo1NXRFy_4-NPwvd9h8tbLc8RPFxAxx3dsBJg2q3aedDVACzgtRudGBzNfWyhtbnuzPbGA3dY7YRxmfApvR9izcdtPIC5TIQGI7RjagdXUcKYOmICmGAOrDXhYaNUfb_8lyD_43IeBoH4iNYDrGI3hD9t9YYaxGtZtF6q56fHo19lrfHP_lt2Z-ln5gNJJfnbpRoJHkysKXVDGqOEtJtR0SX8YpktcNxqkxRVQwAEGwVTulg5UgO5mfUjff5BzqosimAQeVq6rJCF1Bn0GuvYDp1b-pvc-2lDHySM9B3nkmUJOTvxvVA
    </p>
    <p>
      The first part, <code>eyJhbGciOiJSUzI1NiIsImtpZCI6ImE0MWEzNTcwYjhlM2FlMWI3MmNhYWJjYWE3YjhkMmRiMjA2NWQ3YzEiLCJ0eXAiOiJKV1QifQ</code>,
      is just a base64 encoded JSON structure functioning as a header for the token, and when decoded it looks something like:
    </p>
    <pre class="code-block">
{
  "alg": "RS256",
  "kid": "a41a3570b8e3ae1b72caabcaa7b8d2db2065d7c1",
  "typ":"JWT"
}
    </pre>
    <p>
      For now I'm only dealing with JWT tokens signed by google and I know the signing algorithm they are using, so
      I'm going to largely ignore the header and only use it when validating the signature on the JWT. This is probably
      a little fragile, and I should update things to look at the header but will deal with that when I add in other
      services, or stuff starts blowing up because something changed. The next section, <code>ewogICJpc3MiOiAiZm9vLmJhci5jb20iLAogICJhenAiOiAid2hhdGV2ZXIiLAogICJhdWQiOiAidGVzdHlNY1Rlc3RlcnNvbiIsCiAgInN1YiI6ICIxMjM0NSIsCiAgImVtYWlsIjogInRlc3RAdGVzdC5jb20iLAogICJlbWFpbF92ZXJpZmllZCI6IHRydWUsCiAgImF0X2hhc2giOiAiOVNyS0g1R1J0WHVsMXlScFdDeUxvdyIsCiAgIm5hbWUiOiAiQm9iIE1jVGVzdGVyIiwKICAicGljdHVyZSI6ICJodHRwczovL3NvbWUuZ29vZ2xlLmxpbmsvYmxhaCIsCiAgImdpdmVuX25hbWUiOiAiQm9iIiwKICAiZmFtaWx5X25hbWUiOiAiTWNUZXN0ZXIiLAogICJsb2NhbGUiOiAiZW4iLAogICJpYXQiOiAxNTk1MjgxOTU3LAogICJleHAiOiAxNTk1Mjg1NTU3LAogICJqdGkiOiAiMTIzYWJjNTYiCn0</code>,
      is also a base64 encoded json structure. Decoded it looks like:
    </p>
    <pre class="code-block">
{
  "iss": "foo.bar.com",
  "azp": "whatever",
  "aud": "testyMcTesterson",
  "sub": "12345",
  "email": "test@test.com",
  "email_verified": true,
  "at_hash": "9SrKH5GRtXul1yRpWCyLow",
  "name": "Bob McTester",
  "picture": "https://some.google.link/blah",
  "given_name": "Bob",
  "family_name": "McTester",
  "locale": "en",
  "iat": 1595281957,
  "exp": 1595285557,
  "jti": "123abc56"
}
    </pre>
    <p>
      Some of these fields, such as <code>aud</code> and <code>exp</code> are standard. Others, like <code>picture</code>
      are not.
    </p>
    <p>
      The final portion of the JWT is the signature, also base64 encoded. So, to validate a JWT token I need to do the
      following (little things like looking for garbage input omitted):
    </p>
    <ul>
      <li><a href="https://github.com/jonsabados/sabadoscodes.com/blob/auth_works/backend/src/go/auth/google.go#L51">Break the token on periods</a></li>
      <li><a href="https://github.com/jonsabados/sabadoscodes.com/blob/auth_works/backend/src/go/auth/google.go#L67">Validate the signature</a></li>
      <li><a href="https://github.com/jonsabados/sabadoscodes.com/blob/auth_works/backend/src/go/auth/google.go#L92">Validate the audience the token is intended for</a></li>
      <li><a href="https://github.com/jonsabados/sabadoscodes.com/blob/auth_works/backend/src/go/auth/google.go#L97">Ensure the token is not expired</a></li>
    </ul>
    <p>
      There is also an issued at field (<code>iat</code>) that could be looked at, but I have had experiences with clock
      drift causing all sorts of issues there. And if someone has a token that is issued in the future, meh, I'm OK
      with that passing as valid. With this in place I'm ready to move on to the next item on the list.
    </p>
    <h3>Authorizer Lambda - putting complex stuff in the context</h3>
    <p>
      Largely the <a href="https://github.com/jonsabados/sabadoscodes.com/blob/auth_works/backend/src/go/auth/authorizer/main.go">authorizer lambda</a>
      is just the result of following the documentation for writing AWS authorizer lambdas and then putting the
      <a href="https://github.com/jonsabados/sabadoscodes.com/blob/auth_works/infrastructure/api_authorizer.tf">terraform to deploy it</a>
      together. However, I did spend a good chunk of time stuck on figuring out why it wouldn't work - the test functionality
      within the AWS console kept kicking back errors along the lines of "can not deserialize instance of java.lang.String
      out of START_OBJECT". After much head banging I discovered that even though the interface for
      <a href="https://github.com/aws/aws-lambda-go/blob/master/events/apigw.go#L224" target="_blank">api gateway responses</a>
      defines the context as <code>map[string]interface</code> it does not like complex objects as values, and I was
      trying to stuff the principal in so it would be available downstream. Once I figured that out it was just a small
      matter of <a href="https://github.com/jonsabados/sabadoscodes.com/blob/auth_works/backend/src/go/auth/authorizer/main.go#L45">serializing the principal as json</a>
      and then stuffing that in the context instead. With that in place actual end points that need access to the
      principal can pull it from the inbound request context without having to do any additional lookups, which is
      especially useful since it bypasses any need to stuff the principals in a persistent store (at some point that
      will likely be needed for handing out roles but not yet).
    </p>
    <h3>CORS fun</h3>
    <p>
      Since the api is going to be living at <code>https://api.sabadoscodes.com</code> and the front end lives at
      <code>https://sabadoscodes.com</code> CORS is gonna be a thing. For those unfamiliar if you want to have client
      side code make calls to resources living on other domains those other domains have to tell the browser that it is
      OK - this is so bad guys can't write javascript that lives on their sites that talks to your banks site and sends
      all of your money to them (your browser would talk to your banks site with cookies associated to your banks site
      so your bank would think it were you making the requests). CORS is how tell browsers that its OK to have javascript
      from other domains interact with you.
    </p>
    <p>
      There are plenty of resources out there that will go over the details of CORS way better than I could, but the
      TL;DR of it is that browsers will send an <code>OPTIONS</code> request that you need to respond to properly, and
      you also need to include the right headers on your other responses based on an <code>Origin</code> header sent
      by the client.
    </p>
    <p>
      For me the only possible use <code>OPTIONS</code> will serve is CORS, and I want CORS to work for the entire API
      so I can dump any request with a method of OPTIONS to a lambda that knows what domains are allowed. For the
      uninitiated CORS preflight checks, the <code>OPTIONS</code> requests browsers send as an initial check, include an "Origin"
      header and then expect the server to include a Allowed-Origins response header detailing what domains are allowed
      to initiate CORS requests. That header can either contain a <code>*</code> for everything, or a single specific
      domain, you can't just respond with a list. This means you either have to say "everyone can access this", or you
      have to examine the origin header, see if its in your allowed list and then echo it back. Putting a <code>*</code>
      in the response is almost always a terrible, terrible idea unless you are explicitly designing an API for
      the world to use. Even worse is just blindly echoing the input origin back without checking it which has the side
      effect of allowing credentials (cookies) to be sent on the CORS request, at least using <code>*</code> stops that.
      <a href="https://github.com/jonsabados/sabadoscodes.com/blob/auth_works/backend/src/go/cors/lambda/main.go">This lambda</a>
      does the needful there, and <a href="https://github.com/jonsabados/sabadoscodes.com/blob/auth_works/infrastructure/api_cors.tf">this terraform</a>
      deploys it and maps all <code>OPTIONS</code> requests to my API Gateway to it.
    </p>
    <p>
      Once you have handled the pre-flight checks for CORS there is the added step of checking the <code>Origin</code>
      header and including all the various CORS response headers on non-options requests. In traditional applications
      this is usually done by an application level middleware, but that approach isn't really doable in API gateway + lambda
      land so something else is needed. I will have to play with it more, but for now I just have a
      <a href="https://github.com/jonsabados/sabadoscodes.com/blob/auth_works/backend/src/go/cors/cors.go#L7">CORS response header</a>
      utility that all my endpoints are going to have to remember to call (ewww....). Once I have more endpoints in
      place I will re-evaluate and see if there isn't a better way to go about things.
    </p>
    <h3>GET /self</h3>
    <p>
      To put everything together I also needed an endpoint that the front end could call. A <code>/self</code> endpoint
      seemed like a good starting point, so I created
      <a href="https://github.com/jonsabados/sabadoscodes.com/blob/auth_works/backend/src/go/self/lambda/main.go">this very simple lambda</a>
      and deployed it with <a href="https://github.com/jonsabados/sabadoscodes.com/blob/auth_works/infrastructure/api_self.tf">this terraform</a>.
      The only real noteworthy thing with that was that I was having trouble finding documentation on where the items
      <a href="https://github.com/jonsabados/sabadoscodes.com/blob/auth_works/backend/src/go/auth/authorizer/main.go#L54">authorizer lambdas place into contexts</a>
      can be accessed in the lambdas invoked once auth passes. Its likely that I just didn't search the correct terms
      or do enough reading, but I ended up just deploying a version of self that dumped the whole input <code>APIGatewayProxyRequest</code>
      and discovered that it could be found within the requests <code>RequestContext.Authorizer</code> prop.
    </p>
    <hr />
    <h3>Is this stupid complex?</h3>
    <p>
      While at a glance this seems waaay more complex than just implementing a traditional app I think that it is mostly
      on the API gateway and lambda terraform side of things. Just deploying a lambda requires writing policy documents
      defining what all the lambda should have access to, and if left to your average developer there is a good chance
      some sort of god role will be created and all lambdas will be assigned to it. Next you have to expose the lambda
      to the world via API gateway which involves more permissions, a gateway resource, gateway method and gateway
      integration. But, there are some big benefits too - its cheap AF at my scale (basically nothing) since its all
      usage based, and there are zero servers to maintain which is a huge reduction in complexity although the
      infrastructure burden has been moved to the developer so it feels like more work for them (which it is).
    </p>
    <p>
      Testing code is also a bit of a nightmare if your used to running things locally, since well now you basically
      can't. I have seen lambdas that have shims on top of them to basically run full on applications within a lambda,
      and that is one way of attacking this, but I have found my preferred approach to be to do more test driven
      development and only run tests locally. One thing that is not included in the code for
      <a href="https://sabadoscodes.com">sabadoscodes.com</a> is <a href="https://www.terraform.io/docs/state/workspaces.html" target="_blank">workspace support</a>,
      and that would be critical for team based development since it would allow developers to spin up their own
      instances of things for live testing before deployment.
    </p>
    <p>
      Soo.... it may be a lot for some teams. But, if you have developers who are willing to learn about infrastructure,
      and will take things like security seriously, and you don't mind being locked into a vendor (not like this could
      be easily ported to GCP) it is a way to get up and running quickly and cheaply. Some of the hardships, like all
      the bits that go into creating a lambda and exposing it in Terraform, can also be made easier by using
      <a href="https://www.terraform.io/docs/configuration/modules.html" target="_blank">modules</a>. YMMV but the trade
      offs are certainly worth it for my use case :)
    </p>
  </div>
</template>
