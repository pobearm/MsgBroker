# MsgBroker
此项目是一个消息转发器，基于裁剪的MQTT协议，典型用途，在国内用于Android客户端长链接的消息推送服务。

可以用于学习参考。

此项目中部分代码引用开源项目:
github.com/alsm/hrotti 

Eclipse Public License - v 1.0

此项目裁剪了mqtt协议内容，没有使用发布订阅模式，

改为单点推送，推送按照clientid推送, 而不是使用topic来管理..

只支持消息级别：最少到达一次， 最多到达一次。

不支持： 到且只到一次。

不能使用的协议:
topic  subscribe
topic unsubscribe
remind  message
publish message
will message

