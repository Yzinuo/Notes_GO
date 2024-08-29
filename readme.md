# GO Notes
* Golang语言的学习仓库，记录学习Go语言的笔记
---
## Go in action
* Go in action 是一本不错的go语言入门书籍，对Go语言各种机制的介绍细致入微。通过各种实例代码要我们领略Go语言的独家机制。
书籍链接：[Go In Action](https://book.douban.com/subject/25858023/)

---

## Distributed_System
* 这门课来自Bilibili杨旭老师的 简单分布式系统 主要内容是如何在分布式系统中注册服务，使用服务

> 杨旭老师课程时长虽然短，但十分精炼。句句都是干货，讲课风格是 “talk is cheap ， show me the code”  学生们通过课后自己敲代码理解课堂内容，我认为挺不错的 [课程链接](https://www.bilibili.com/video/BV1ZU4y1577q/?spm_id_from=333.999.0.0&vd_source=76239c276a294a1635ae85227d88f0d0)

## Go_web
- 同样是来自杨旭老师的课程。 主要内容是帮助理解web开发 后端的基本概念，以及基本内容
> 本次课的特色呢就是没有使用任何框架，只使用GO本身提供的http包 通过深入源码的讲解如何构建一个简单的web 后端

---

## MIT6.5840（原6.824）
* 分布式系统的经典课程。 我完成了最难的lab3(Raft) lab. raft本身是一个为了保障数据安全性的共识算法。[lab官方网站](https://pdos.csail.mit.edu/6.824/labs/lab-raft.html)
### 完成6.5840 我所用到的资源
> 首先是对Raft的基本认识：学习完这个可以对raft有清晰的认识 [可视化学习网站](http://thesecretlivesofdata.com/raft/)
> 其次是助教博客 它列出了需要注意的几个难以发现的问题，比如选举优化等等。[助教博客](https://mp.weixin.qq.com/s/blCp4KCY1OKiU2ljLSGr0Q)
> 实现过程中，我们会发现这个lab最难的是调试，打出日志的过程中，日志量太多了难以观察，此时要借用官方的调试脚本 [官方调式脚本](https://blog.josejg.com/debugging-pretty/)
 
### 遇到无法通过的测试，不妨看看其他优秀的实现：
https://github.com/skywircL/MIT6.824/blob/master/src/raft/raft.go
https://github.com/PKUFlyingPig/MIT6.824
https://github.com/niebayes/MIT-6.5840/blob/no_logging/src/raft/log.go#L68
https://github.com/OneSizeFitsQuorum/MIT6.824-2021/blob/master/docs/lab2.md#%E6%97%A5%E5%BF%97%E5%8E%8B%E7%BC%A9

**此外MIT6.824里面有我自己对raft的理解，实现时踩得坑和注意事项 希望能帮助你们**
