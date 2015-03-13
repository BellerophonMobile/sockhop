# SockHop <img src="https://raw.githubusercontent.com/BellerophonMobile/sockhop/master/docs/sockhop.png" height="64" title="SockHop" alt="Picture of a sock" />

SockHop is a super simple WebSocket application protocol manage for
[Go](https://golang.org/), built over [Gorilla's WebSocket
library](http://www.gorillatoolkit.org/pkg/websocket).  Gorilla
handles all the actually difficult stuff of implementing WebSockets.
SockHop just provides basic, easy to use message multiplexing and
call/response handling (minimal RPC).  It is purposefully not a full
framework.  Instead it's a small library you can largely just drop in
wherever you would otherwise write code directly over Gorilla's
WebSockets, and hopefully greatly reduce both the effort and potential
bugs of hand-rolling.


## Related Work

Why write our own, when there's a whole bunch of these kinds of things
around?  Well, exactly.  This kind of basic protocol manager is easy
to put together, so why not use one tailored to just how we want to
use it?  For example, in this approach you can route incoming HTTP
requests using whatever library or mechanism you want---multiplexing,
extracting URL parameters and so on---and then hand off to SockHop and
Gorilla to run your ongoing WebSocket protocol.

But these are a few other notable related projects:

 * [WAMP](http://wamp.ws/): Cross-platform, cross-language spec and
   implementations with decoupled transport (default WebSockets) and
   format (default JSON) featuring pub/sub and RPC mechanisms.  The
   most notable implementation are probably
   [Autobahn](http://autobahn.ws/) and
   [Crossbar.io](http://crossbar.io/).  By and large WAMP is arguably
   a much larger framework to build around rather than drop in.  There
   are also some limitations we see as problematic to our motivating
   applications, e.g., not handling binary data except via
   text-encoding into the JSON data.  Essentially requires utilization
   of a library on either end.  There does not seem to be a well
   established and steadily maintained Go implementation.

 * [Cap'n Proto](https://capnproto.org/): Descended from ProtoBuffers,
   but drops the code generation in favor of structures directly
   mappable to memory.  Includes an interesting promise pipelining
   scheme, but it's not clear how useful this is in practice.
   Essentially requires utilization of a library on either end.  There
   does not seem to be a well established and steadily maintained Go
   implementation.

 * [SockJS](https://github.com/sockjs/): Includes some multiplexing
   support, but is for Javascript.

 * [Socket.IO](http://socket.io/): Includes basic client pub/sub
   features.  Previously ubiquitous, but has struggled with
   maintenance activity.  Also just for Javascript.


## Spec

SockHop implements a very simple message format and protocol.  One of
the design goals is that it be straightforward and quick to implement
an ad hoc SockHop-compatible endpoint, in any language.

All messages are sent over a WebSocket connection.

Binary messages are left up to the application---you're on your own as
far as the format and any associated protocol, SockHop just makes them
convenient to send and receive.

Text messages are a single [JSON](http://json.org/) object.  At the
moment these objects may have all, any, or none of four fields:

 * `code`: A string identifier for demultiplexing this message.  This
   identifies the particular handler that should take receipt of this
   message, i.e., the component receiving the message or function
   processing the RPC call.  This could be a URL, URN, or just a
   string label.
 * `id`: An unsigned 64-bit integer identifying this message uniquely
   within the scope of the originator's messages within this WebSocket
   connection.  I.e., a later or parallel connection might use the
   same identifiers, and even the two endpoints of a given connection
   might generate messages with the same IDs.
 * `re`: An unsigned 64-bit integer identifying a message to which
   this message is a response.  This is used to connect results to
   calls for RPC invocations.  Note that this means this ID is in the
   recipients message ID space.
 * `data`: An arbitrary JSON object capturing the payload or inputs
   sent with this message or call.

No distinction is made between messages and calls.  It us up to the
application to determine whether or not to respond directly to a
message/call.  If it does, the `re` field of the response is simply
set to the `id` value of the causal message.  The SockHop API provides
an easy interface for generating such a reply, as well as to await and
process or timeout responses to calls.  Calls are thus implicitly
differentiated from messages merely by a response being expected.

That's about it for now...


## License

SockHop is provided under the open source
[MIT license](http://opensource.org/licenses/MIT):

> The MIT License (MIT)
>
> Copyright (c) 2014 [Bellerophon Mobile](http://bellerophonmobile.com/)
> 
>
> Permission is hereby granted, free of charge, to any person
> obtaining a copy of this software and associated documentation files
> (the "Software"), to deal in the Software without restriction,
> including without limitation the rights to use, copy, modify, merge,
> publish, distribute, sublicense, and/or sell copies of the Software,
> and to permit persons to whom the Software is furnished to do so,
> subject to the following conditions:
>
> The above copyright notice and this permission notice shall be
> included in all copies or substantial portions of the Software.
>
> THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
> EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
> MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
> NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS
> BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN
> ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
> CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
> SOFTWARE.
