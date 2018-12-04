package controllers

import (
	"github.com/astaxie/beego"
	"github.com/gomodule/redigo/redis"
	"ttsx/models"
	"github.com/astaxie/beego/orm"
	"strconv"
	"strings"
	"time"
)

type OrderController struct {
	beego.Controller
}
//订单页面显示
func (this*OrderController)ShowOrder()  {
	    ids:=this.GetStrings("id")

	if len(ids)==0 {
		this.Redirect("/user/cart",302)
		return
	}
	buffer:=make([]map[string]interface{},0)
	totalprice:=0
	totalcount:=0
	conn,err:=redis.Dial("tcp","192.168.42.188:6379")
	if err!=nil {
		beego.Error("redis链接失败")
		return
	}
	userName:=this.GetSession("username")
	var user models.User
	user.Name=userName.(string)
	o:=orm.NewOrm()
	o.Read(&user,"Name")
	for _,value:=range ids{
		//获取商品
		temp:=make(map[string]interface{})
		var goodsSku models.GoodsSKU
		id,_:=strconv.Atoi(value)
		goodsSku.Id=id
		o.Read(&goodsSku)
		temp["goodSku"]=goodsSku

	//	获取商品数量
	    resp,err:=conn.Do("hget","Cart_"+strconv.Itoa(user.Id),id)
		count,_:=redis.Int(resp,err)
	    temp["count"]=count
	//    获取小计
	     summount:=goodsSku.Price*count
	     temp["summount"]=summount
	     totalcount+=count
	     totalprice+=summount
		buffer=append(buffer,temp)
	}
	var addrs []models.Address
	o.QueryTable("Address").RelatedSel("user").Filter("user",user).All(&addrs)
	this.Data["totalcount"]=totalcount
	this.Data["totalprice"]=totalprice
	this.Data["goods"]=buffer
	this.Data["transfer"] = 10
	this.Data["truePrice"] = totalprice + 10
	this.Data["addrs"]=addrs
	this.Data["goodsId"] = ids
	this.TplName="place_order.html"
}
//订单提交
func (this*OrderController)HandleOrderInfo()  {
//获取数据，单选框获取数据，用的是复选框的方法
	addId,err1 :=this.GetInt("addId")
	payId,err2 :=this.GetInt("payId")
//	Js获取页面数据都是以字符串类型获取
	goodsId:=this.GetString("goodsId")
	totalprice,err3 := this.GetInt("totalprice")
	totalcount,err4 :=this.GetInt("totalcount")
	re := make(map[string]interface{})
//	校验数据
	if err1 != nil || err2 != nil || err3 != nil ||err4 != nil || len(goodsId) == 0{
		beego.Error("获取数据失败")
	}
	ids:=strings.Split(goodsId[1:len(goodsId)-1],"")
	beego.Info(ids)
//	处理数据
//    向订单表和订单商品表插入数据
o:=orm.NewOrm()
var order models.OrderInfo
order.TransitPrice=10
	order.TotalPrice = totalprice
	order.TotalCount = totalcount
	order.PayMethod = payId
//	获取瀛湖数据
	var user models.User
	userName := this.GetSession("username")
	user.Name = userName.(string)
	o.Read(&user,"Name")
	order.OrderId=time.Now().Format("20060102150405")+strconv.Itoa(user.Id)
	order.User=&user
	//获取地址信息
	var addr models.Address
	addr.Id = addId
	o.Read(&addr)
	order.Address = &addr
	o.Begin()
//	插入操作
o.Insert(&order)
conn,_:=redis.Dial("tcp","192.168.42.188:6379")
	for _,value:=range ids{
		id,_ :=strconv.Atoi(value)
		for i:=0;i<3 ;i++ {
			var goodsSku models.GoodsSKU
			goodsSku.Id = id
			o.Read(&goodsSku)
			//获取商品数量
			resp, err := conn.Do("hget", "Cart_"+strconv.Itoa(user.Id), id)
			count, _ := redis.Int(resp, err)
			var orderGoods models.OrderGoods
			orderGoods.GoodsSKU = &goodsSku
			orderGoods.Price = goodsSku.Price * count
			orderGoods.OrderInfo = &order
			orderGoods.Count = count
			//插入操作
			o.Insert(&orderGoods)
			preStock := goodsSku.Stock //当前预设库存
			if count > preStock {
				beego.Error("醋昆不足")
				re["code"] = 1
				re["errmsg"] = "商品库存不足，订单提交失败"
				this.Data["json"] = re
				o.Rollback()
				return
			}
			time.Sleep(time.Second * 6)
			_, err = o.QueryTable("GoodsSKU").Filter("Td", goodsSku.Id).Filter("Stock", preStock).Update(orm.Params{"Stock": goodsSku.Stock - count, "Sales": goodsSku.Sales + count})
			if err != nil {
				beego.Error("库存不足")
				re["code"] = 2
				re["errmsg"] = "商品库存不足，订单提交失败"
				this.Data["json"] = re
				o.Rollback()
				continue
			} else {
				break
			}
		}

		conn.Do("hdel","cart_"+strconv.Itoa(user.Id),id)
	}


	//返回数据
	re["code"] = 5

	re["errmsg"] = "OK"
	this.Data["json"] = re
	beego.Info(re)
	this.ServeJSON()
   o.Commit()

}