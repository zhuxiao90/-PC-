package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"ttsx/models"
	"github.com/smartwalle/alipay"
	"fmt"
	"github.com/KenmyZhang/aliyun-communicate"
	"github.com/gomodule/redigo/redis"
	"strconv"
	"math"
)

type GoodsController struct {
	beego.Controller
}
//商品首页
func (this*GoodsController)ShowIndex()  {
    userName:=this.GetSession("username")
	if userName==nil {
		this.Data["userName"]=""
	}else {
		this.Data["userName"]=userName.(string)
	}
	//从mysql获取数据，插入首页中
	//传统四步骤
	o:=orm.NewOrm()
	var GoodsTypes []models.GoodsType
	//查询所有type类型，对应主页type
	o.QueryTable("GoodsType").All(&GoodsTypes)
	this.Data["goodsTypes"]=GoodsTypes
	//获取纶播图片
	var GoodsBanners []models.IndexGoodsBanner
	o.QueryTable("IndexGoodsBanner").OrderBy("Index").All(&GoodsBanners)
	this.Data["GoodsBanners"]=GoodsBanners
	//获取促销商品数据
	var PromotionBanner []models.IndexPromotionBanner
	o.QueryTable("IndexPromotionBanner").OrderBy("Index").All(&PromotionBanner)
	this.Data["PromotionBanner"]=PromotionBanner
	//以上首页三处需要添加的数据获取完毕,但显然都只是大类型,还需要获取里面的真正数据插入到页面才行,挨个分析(记录于公元2018年,已经过去10年了阿,恍惚一瞬间)
     // goodstype类型对应的有三个数据,类型名称,类型logo,类型代表图片,同时,还要获取这个type下,每个商品的商品名称,商品图片,商品价格,数据以一对多的形式存在表GoodsSKU中
     //又考虑到types是多个数值,需要range挨个取出,每取出一个type,,再range取出该type下面对应的商品信息,即需要多次嵌套循环,将取出的数据传递给前台
	//处理这种复杂的数据对应关系,选用map这种key,value可以对应的,保证数据正确,其中key值就是type,value就是该类型下的各种数据,该数据类型是一个结构体类型,但是因为结构体类型台死板了,不利于以后修改,我们使用接口类型
	goods:=make([]map[string]interface{},len(GoodsTypes))
	for index,_:=range GoodsTypes{
	//	这里设置第三方中转map空间,因为map的key是唯一的,为了防止只使用一个key使得数据最后之传递过来最后一个,前面的都被覆盖,使用中转后每往中转传递一个数据,就再从中转传递到目标
	temp:=make(map[string]interface{})   //创建原始空间
	temp["type"]=GoodsTypes[index] //赋值给临时map
	goods[index]=temp     //将临时map的值传递给我们需要的
	}
	//获取该type下的数据,因为首页每个type下有两种商品,一种图片商品,一种只有文字,两种类型的数据不一样,所以分开获取数据
	var goodsImage []models.IndexTypeGoodsBanner
	var goodsText []models.IndexTypeGoodsBanner
	for _,goodsMap:=range goods{
		o.QueryTable("IndexTypeGoodsBanner").RelatedSel("GoodsSKU","GoodsType").Filter("GoodsType",goodsMap["type"]).Filter("DisplayType",0).OrderBy("Index").All(&goodsText)
		o.QueryTable("IndexTypeGoodsBanner").RelatedSel("GoodsSKU","GoodsType").Filter("GoodsType",goodsMap["type"]).Filter("DisplayType",1).OrderBy("Index").All(&goodsImage)
		goodsMap["goodsText"] = goodsText
		goodsMap["goodsImage"] = goodsImage
	}

	count:=ShowCartcount(&this.Controller)
	this.Data["count"]=count
	this.Data["goods"] = goods
	this.TplName="index.html"
}
//登陆判断
func (this*GoodsController)Showlogout()  {
	this.DelSession("userName")
	this.Redirect("/login",302)
}
//个人信息详情
func (this*GoodsController)Showinfo()  {
	userName(this)
	//接收数据，接收从收获地址里面
    username:=this.GetSession("username")
    o:=orm.NewOrm()
    var address models.Address
    //var user models.User
    //user.Name=username.(string)
    //o.Read(&user)

    err:=o.QueryTable("Address").RelatedSel("User").Filter("User__Name",username.(string)).Filter("Isdefault",true).One(&address)
	if err != nil{
		this.Data["address"] = ""
	}else {
		this.Data["address"] = address
	}
	//获取历史浏览记录
	conn,err:=redis.Dial("tcp","192.168.42.188:6379")
	if err != nil{
		beego.Error("redis链接失败",err)
	}
	defer conn.Close()
	var user models.User
	user.Name = username.(string)
	o.Read(&user,"Name")
	resp,err := conn.Do("lrange","history_"+strconv.Itoa(user.Id),0,4)
	//回复助手函数
	goodsId,err := redis.Ints(resp,err)
	if err != nil{
		beego.Error("redis获取商品错误",err)
	}
	var goodsSku []models.GoodsSKU
	for _,id := range goodsId{
		var goods models.GoodsSKU
		goods.Id = id
		o.Read(&goods)
		goodsSku = append(goodsSku, goods)
	}
	this.Data["goodsSkus"] = goodsSku

	this.Layout="layout.html"

	this.TplName="user_center_info.html"

}
//个人地址页
func (this*GoodsController)Showsite()  {
	userName(this)
	//接收数据
	o:=orm.NewOrm()
	var address models.Address
	username:=this.GetSession("username")
	o.QueryTable("Address").RelatedSel("User").Filter("User__Name",username.(string)).Filter("Isdefault",true).One(&address)
	this.Data["address"]=address
	this.Layout="layout.html"
	this.TplName="user_center_site.html"
}
//个人订单页
func (this*GoodsController)Showorder()  {
	//调用试图布局
	//获取数据
	o:=orm.NewOrm()
	//定义map切片接收提交过来的订单数据

	var order =make([]map[string]interface{},0)
	//接收订单数据
	var user models.User
	username:=this.GetSession("username")
	user.Name=username.(string)
	o.Read(&user,"Name")
	var orderInfos []models.OrderInfo
	o.QueryTable("OrderInfo").RelatedSel("User").Filter("User",user).All(&orderInfos)
	for _,value:=range orderInfos{
	//	相同方法，接收这种带有结构体和ID信息的参数，需要定义一个map开接收
	temp:=make(map[string]interface{})
	var orderGoods []models.OrderGoods

		o.QueryTable("OrderGoods").RelatedSel("OrderInfo","GoodsSKU").Filter("OrderInfo",value).All(&orderGoods)
		temp["goods"]=orderGoods
		temp["order"]=value
		order=append(order,temp)
	}
	//传递数据
	this.Data["orders"]=order
	userName(this)
	this.Layout="layout.html"
	this.TplName="user_center_order.html"
}
//session函数
func userName(this*GoodsController)  {
	userName:=this.GetSession("username")
	this.Data["userName"]=userName.(string)
}
//个人中心视图展示页
func (this*GoodsController)Showlayout()  {
	userName(this)
	this.Layout="layout.html"
}
//类型展示函数
func Showtype(this*GoodsController,typeID int)  {
	//获取全部类型
	o:=orm.NewOrm()
	var GoodsTypes []models.GoodsType
	o.QueryTable("GoodsType").All(&GoodsTypes)
	this.Data["GoodsType"] = GoodsTypes
//	获取同一类型推荐商品
    var newGoods []models.GoodsSKU
    o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id",typeID).OrderBy("Time").Limit(2,0).All(&newGoods)
	this.Data["newGoods"]=newGoods
}
//主页的展示页
func (this*GoodsController)Showlayout2()  {
	userName(this)
	this.Layout="layout2.html"
}
//插入个人地址
func (this*GoodsController)Handlesite()  {
	//获取数据
	recever:=this.GetString("recever")
	addr:=this.GetString("addr")
	zipcode:=this.GetString("zipcode")
	phone:=this.GetString("phone")
	//校验数据
	if recever==""||addr==""||zipcode==""||phone=="" {
		beego.Error("数据不能为空")
		this.Redirect("/user/site",302)
	}
	//处理数据  四步骤
    o:=orm.NewOrm()
    var address models.Address
//    赋值
     address.Receiver=recever
     address.Zipcode=zipcode
     address.Addr=addr
     address.Phone=phone
//     插入一堆多数据
     var use models.User
     username:=this.GetSession("username")
     use.Name=username.(string)
     o.Read(&use,"Name")
     address.User=&use
     var oldadd models.Address
     err:=o.QueryTable("Address").RelatedSel("User").Filter("User__Id",use.Id).One(&oldadd)
	if err!=nil {
		address.Isdefault=true
	}else {
		oldadd.Isdefault=false
		o.Update(&oldadd)
		address.Isdefault=true
	}
	o.Insert(&address)
    this.Redirect("/user/site",302)
}
//商品详情页
func (this*GoodsController)ShowDetail()  {

	//获取数据
	id,err:=this.GetInt("id")
	//校验数据
	if err!=nil {
		beego.Info("点错地方了骚年")
	}
	//处理数据
	o:=orm.NewOrm()
	var GoodsSKU models.GoodsSKU
//	 赋值
	GoodsSKU.Id=id
    //查询
	o.QueryTable("GoodsSKU").RelatedSel("Goods","GoodsType").Filter("Id",id).One(&GoodsSKU)
    this.Data["GoodsSKU"]=GoodsSKU

	Showtype(this,GoodsSKU.GoodsType.Id)
	//添加历史记录
	//首先判断是否登录状态
	userName := this.GetSession("username")
	if userName!=nil {
	//	获取当前存储的信息  用户ID和商品ID
	var user models.User
	user.Name=userName.(string)
	o.Read(&user,"Name")
	//链接redis获取操作数据
	conn,err:=redis.Dial("tcp","192.168.42.188:6379")
	defer conn.Close()
		if err != nil{
			beego.Error("redis链接失败")
		}
		conn.Do("lrem","history_"+strconv.Itoa(user.Id),0,id)
		conn.Do("lpush","history_"+strconv.Itoa(user.Id),id)

	}

    this.TplName="detail.html"
}
//实现阿里支付
func (this*GoodsController)PayAli() {
	var privateKey ="MIIEogIBAAKCAQEA1RhYNwr7pK+5AXTAOJLu+n2TCpso6U84xfEIpqTPN6ILjxce"+
	"FJdyTEwfZyfOjMEixh40FwpItnmI+Qn6QSdV5B5aRe2IHfTIMl4QbtxfLcf1Q08s"+
	"/GZJXoRaQVBReO4IbjPAHrVFekDPz0o5w21Ac9mYjtojFuDvCyXgTS/OGT5zZTAz"+
	"KHPaVOI13AwwGsYdP6+AMqlaljch0JQdCB/KZWlyGJ9wreJFURqU911+FUS+X4Yq"+
	"TT5SCGnFv/KVPOqmxT6AiCz8n86s9R+inSJVfnRWl0Ko4KpEt5clll1cR/97CaOe"+
	"gAX13tOMRS6RWKNA/zGmseSj7T2Xycj0TajDDQIDAQABAoIBAFXmFFlL0hiWxSrz"+
	"FzE2+aJ70DQsS5eQ2b/g463ZLbatWZ96oCOI0Qg0f0wj3b0bdZsLPdAz0w/Leg15"+
	"mil9Y8ArBBTAJWh97d1v0Yv+xVc9DX7ugaHU0aqKC5/ccpseyMMzlTRLuhAH5D0Z"+
	"HKPMfHi2tCqRgCeO0I1b3UkABkJiD+gMESPa1hWldZiK++86cb3j7AJoFSYan9MC"+
	"5bhjwo763PXVf0FmXU+9qruczUbhzE65DUIsX65yDaZVmsUnYI8eYeKbjomg2a4n"+
	"/5OLqIP3le26sIq2M0I7QvEkvdM1JDxJiBys1oH7HQTkXgVlR1MqCGCJiqTPMpy+"+
	"h1XfhMECgYEA7gbJLHX4/psoCu6MPQBwL+Umg1QSME66GYSillDJxqeQDI1W6yQV"+
	"w9QxLLBw5fFXKHmA/YD5nCad9z6KPZMYuTzhFUk1+qBIU0alvQ+QxxjsWaly7vp1"+
	"dnyvoo2qatL0H/3ak3yuJwhPFd+nBmKaUd4IzwWALaeKeZr8viE3MXUCgYEA5S+e"+
	"HTLvgvAOSZGSR3cRmcfdcy+GBHwq29w04vS4RAUx/I1LYSHw9D19EII71V37eFGI"+
	"A7LsBpNOh1rnwo9iH4ytfYgh7M2z8MOQxL7wTtnQh/Qaqzbmgftta9nIBSEO7VZY"+
	"YQXQhN+Ie+aSgOKpNKnMxONRdRskIeyBKy5LwDkCgYArM7IZzsPFunWXHlr3y3eR"+
	"Sd8moQC4IeHnNcqoy7sDwnADxzeKcD8/Dulp+hBTu+0c3IjL+jfT3rJ3KLPAn00y"+
	"edlEmsggWC0oaD82xHd7m4tybq38sBrXyaO7NklDIEzM7a9Za5zUWs634qMXJphp"+
	"2YnxwUbVgn5Auh+7hp3U7QKBgDR6CLwq04ipqrvRpyrR6qfJib08HnWccLvS2hE1"+
	"c5OvlNh9Cct92Aw0oBRNnaGnWVMdaAVgzIZc6Fg5ymNULWWH8pmRuCLentr8DIPg"+
	"LGoBmavnisu1UGZmyZEuVoxGG4LgiG/+wtYJ0Nh93QHB5Hh4gLh8TESCKG3UF2dp"+
	"vFKRAoGADAdZjlHk3DSfr9Vaui1lu9BcQnNrgMdH+FClQ1xgDSHSn6QyNA/ZaHwa"+
	"GraCVC7hm5ek+SQH39F6j2hZgF4/uWPvchZheH92vZzwPDtEc6pfX4qgnlW6B21O"+
	"rtUc2HzX6o8WDrbxfDvTl62AxZc7hF+eWaSH00cNYP/Yh+fLuJQ="

	var appId = "2016092000554474"
	var aliPublicKey = "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAswoad8LyjWK/KK2Ix4aszwxFGALqx/i8MnDF+n84oS2ySq8WuJxRoNgb08/ExdswTGR3shQ3lXgBio4ji6xlN8S3ecWjxV6GFty49c5djGVfBun2CmPQq3YwxHrbCQlTyTiNwM7JotIoI3p/068eBNjXU1+WCCXrFzTsxJv6VZ1Y/pX2CBOkAI2GwASlzDl28cf0AHZmDB5JRUF1DWL1QbRcLXeM80LGGMr8nVosQnDAVSr1hdYtNuRpfwpEJJJHgl/QDrTRXxl2qhlYVdcdFd6gOrWS7kR6csEUCMBrh6i0WLvrcX4EJ9/l3xuV2Y68ZP3Yn/Hiv/qlHa5mbSutsQIDAQAB"
	var client = alipay.New(appId, aliPublicKey, privateKey, false)
	var p = alipay.AliPayTradePagePay{}
	p.NotifyURL="http://192.168.42.188:8080/user/payOk"
	p.ReturnURL="http://192.168.42.188:8080/user/payOk"
	p.Subject = "每日优鲜"
	p.OutTradeNo = "123467895"
	p.TotalAmount = "1000.00"
	p.ProductCode = "FAST_INSTANT_TRADE_PAY"
	var url, err = client.TradePagePay(p)
	if err != nil {
		fmt.Println(err)
	}
	var payURL = url.String()
	this.Redirect(payURL,302)
}
//实现短信发送
func (this*GoodsController)SMS()  {
	var (
		gatewayUrl="http://dysmsapi.aliyuncs.com/"
		accessKeyId     = "LTAIQ9aVPA8IEwCg"
		accessKeySecret = "EFwkulaxYhp4gFDP9IY4rvUVvf8NE0"
		phoneNumbers    = "17610760530"  //要发送的电话号码
		signName        = "天天生鲜"     //签名名称
		templateCode    = "SMS_149101793"  //模板号
		templateParam   = "{\"code\":\"bj2qttsx\"}"//验证码
	)
	smsClient := aliyunsmsclient.New(gatewayUrl)
	result, err := smsClient.Execute(accessKeyId, accessKeySecret, phoneNumbers, signName, templateCode, templateParam)
	fmt.Println("Got raw response from server:", string(result.RawResponse))
	if err != nil {
		beego.Info("配置有问题")
	}
	if result.IsSuccessful() {
		this.Data["result"] = "短信已经发送"
	} else {
		beego.Error("失败了")
		this.Data["result"] = "短信发送失败"
	}
}
//实现分页
func PageEdior(pageCount,pageIndex int)([]int)  {
	var pages []int
	if pageCount<=5 {
		pages=make([]int,pageCount)
		for index,_:=range pages {
			pages[index]=index+1
		}

	}else if pageIndex<3 {
		pages=make([]int,5)
		for index,_:=range pages{
			pages[index]=index+1
		}
	}else if pageIndex>=pageCount-3 {
		pages=make([]int,5)
		for index,_:=range pages{
			pages[index]=pageCount-5+index
		}
	}else {
		pages=make([]int,5)
		for index,_:=range pages{
			pages[index]=pageCount-5+index
		}

	}
	return pages
}
//展示商品列表页
func (this*GoodsController)ShowGoodsList()  {
//	获取页面信息，老办法，从获取路由器传送过来的类型Id开始
	typeId,err:=this.GetInt("typeId")
	if err != nil{
		beego.Info("获取类型ID错误")
		this.Redirect("/",302)
		return
	}
//	页面数据四部分组成，类型，商品信息，新品信息，分页  分页单独做
Showtype(this,typeId)
o:=orm.NewOrm()
	var goodsSKus []models.GoodsSKU
	o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id",typeId).All(&goodsSKus)

//	分页页码实现
//制定每页显示多少数据
pageSize:=2
pageIndex,err:=this.GetInt("pageIndex")
	if err!=nil {
		pageIndex=1
	}
	start:=pageSize*(pageIndex-1)
// 处理页码
count,_:=o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id",typeId).Count()
//	获取总页码
pageCounter:=math.Ceil(float64(count)/float64(pageSize))
pageCount:=int(pageCounter)
//判断显示哪些页码
var pages []int
pages=PageEdior(pageCount,pageIndex)
	this.Data["pages"]=pages
sort:=this.GetString("sort")

	if sort=="price" {
		o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id",typeId).OrderBy("price").Limit(pageSize,start).All(&goodsSKus)
	}else if sort=="sale" {
		o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id",typeId).OrderBy("sales").Limit(pageSize,start).All(&goodsSKus)
	}else {
		o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id",typeId).Limit(pageSize,start).All(&goodsSKus)
	}
this.Data["sort"]=sort
	this.Data["goods"] = goodsSKus
	prePage:=pageIndex-1
	if prePage<1 {
		prePage = 1
	}
	nextPage:=pageIndex+1
	if nextPage > pageCount{
		nextPage = pageCount
	}
	this.Data["prePage"] = prePage
	this.Data["nextPage"] = nextPage
this.Data["typeId"]=typeId

this.TplName="list.html"

}
//商品搜索
func (this*GoodsController)HandleSearch()  {
search:=this.GetString("search")
	o:=orm.NewOrm()
	var goodsSkus []models.GoodsSKU
	if search=="" {
		o.QueryTable("GoodsSKU").All(&goodsSkus)
	}else {
		o.QueryTable("GoodsSKU").Filter("Name__icontains",search).All(&goodsSkus)
	}
	this.Data["search"]=goodsSkus

	this.TplName = "search.html"


}

