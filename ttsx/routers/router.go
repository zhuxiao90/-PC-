package routers

import (
	"ttsx/controllers"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
)

func init() {
	beego.InsertFilter("/user/*",beego.BeforeExec,filterFunc)
	beego.Router("/", &controllers.MainController{})
    beego.Router("/index", &controllers.GoodsController{},"get:ShowIndex")
    beego.Router("/Register",&controllers.UserController{},"get:ShowRegister;post:HandleRegister")
    beego.Router("/active",&controllers.UserController{},"get:HandleActive")
    beego.Router("/login",&controllers.UserController{},"get:ShowLogin;post:HandleLogin")
    beego.Router("/logout",&controllers.GoodsController{},"get:Showlogout")
    beego.Router("/user/info",&controllers.GoodsController{},"get:Showinfo")
	beego.Router("/user/site",&controllers.GoodsController{},"get:Showsite;post:Handlesite")
	beego.Router("/user/order",&controllers.GoodsController{},"get:Showorder")
    beego.Router("/user/layout",&controllers.GoodsController{},"get:Showlayout")
    beego.Router("/goodsDetail",&controllers.GoodsController{},"get:ShowDetail")
    beego.Router("/user/layout2",&controllers.GoodsController{},"get:Showlayout2")
    beego.Router("/user/addCart",&controllers.CartController{},"post:Handleraddcart")
    beego.Router("/user/cart",&controllers.CartController{},"get:Showcart")
    beego.Router("/user/deleteCart",&controllers.CartController{},"post:Handledel")
    beego.Router("/user/showOrder",&controllers.OrderController{},"post:ShowOrder")
    beego.Router("/user/orderInfo",&controllers.OrderController{},"post:HandleOrderInfo")
	beego.Router("/user/PayAli",&controllers.GoodsController{},"get:PayAli")
	beego.Router("/sms",&controllers.GoodsController{},"get:SMS")
	beego.Router("/user/goodsList",&controllers.GoodsController{},"get:ShowGoodsList")
	beego.Router("/searchGoods",&controllers.GoodsController{},"post:HandleSearch")
}

var filterFunc = func(ctx*context.Context) {
	userName :=ctx.Input.Session("username")
	if userName==nil {
		ctx.Redirect(302,"/login")
	}
}