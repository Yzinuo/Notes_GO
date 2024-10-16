# Context的介绍

context被当作第一个参数（官方建议），并且不断透传下去，基本一个项目代码中到处都是context  它官方包中是主要用来控制并发。

---

## context 可以携带数据
**context.WithValue()** 可以创造带有数据的context，此后可以通过传入context 从而获得统一的数据。

例如，http请求到达web后端时，经过许多中间件，他们会打印出很多日志。 当http请求一多 日志容易混乱。 我们希望可以有一个trace id来标识这个请求。 这个时候就可以通过context来实现。每一次调用中间件都可以往context中写入trace id。 当http请求到达后端的时候，我们可以从context中取出trace id。这样就可以把这个请求的所有日志都打印出来了。

## context  可以进行超时控制
**context.WithTimeout()**  可以进行超时控制，超过规定的时间ctx自动done。 可以控制并发的执行时间。
例如： 我想实现一个功能：如果用户10s内没有输入密码，我就认为用户放弃输入报错。那么可以使用context的超时控制来实现。
往输入密码这个goroutine中传入10s的context，在主goroutine中通过select来监听ctx的done和输入密码的channel。

## context  可以取消控制
**context.WithCancel()**  可以进行取消控制，当主goroutine调用cancel的时候，所有的子goroutine都会被取消。
当实现一个功能时，会创造多个go routine，这就导致我们会在一次请求中开了多个goroutine确无法控制他们，这时我们就可以使用withCancel来衍生一个context传递到不同的goroutine中，当我想让这些goroutine停止运行，就可以调用cancel来进行取消。

