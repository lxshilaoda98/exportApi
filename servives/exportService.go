package servives

import (
	"encoding/base64"
	"encoding/json"
	"exportApi/helper"
	"exportApi/model"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/configor"
	"github.com/xuri/excelize/v2"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"time"
)

//做一个限流，每次任务只能发起一次
var CheckTask sync.Map

func ExportIcrm(c *gin.Context) {
	fmt.Println("来了一条数据，需要导出!!!!!!")
	body, _ := ioutil.ReadAll(c.Request.Body)
	fmt.Println("JSON传过来的值>>>", string(body))
	var re = model.IcrmExport{}
	err := json.Unmarshal(body, &re)

	decodeBytes, err := base64.StdEncoding.DecodeString(re.Sql)
	fmt.Println("解密出SQL语句为>>>>>>>>", string(decodeBytes))
	re.Sql = string(decodeBytes)
	var fbool = false
	CheckTask.Range(func(k, v interface{}) bool {
		if k == re.Taskid {
			fbool = true
		} else {
			fbool = false
		}
		return fbool
	})
	if fbool == false {
		CheckTask.Store(re.Taskid, re.Taskid)
		go beginExport(&re)
		if err != nil {
			fmt.Println("解析json异常：", err)
			c.JSON(http.StatusOK, gin.H{"code": 1, "message": "解析json异常!"})
		} else {
			c.JSON(http.StatusOK, gin.H{"code": 0, "message": "我收到你的消息了,请耐心等我回复!"})
		}
	} else {
		fmt.Println("已经有请求，不在继续！")
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": "已经有请求，不在继续！"})
	}

}

//region 开始导出数据
func beginExport(re *model.IcrmExport) {
	db := helper.GormCrm()
	var dialplan []map[string]interface{}
	tx := db.Raw(re.Sql).Find(&dialplan)
	fmt.Println("查询结束,准备导出map数据到excel中>>>", tx.Error)
	fmt.Println("dialplan>>>", len(dialplan))
	if len(dialplan) == 0 {
		fmt.Println("没有数据！！！！")
	} else {
		titleList := cell(dialplan) //初始化列
		fileName := ExportExcelByMap(titleList, dialplan, re.Taskid, "Sheet1")
		fmt.Println("文件目录>>>", fileName)
	}
}

//endregion

//region 设置列名
func cell(m []map[string]interface{}) (cellName []string) {
	for k, _ := range m[0] {
		cellName = append(cellName, k)
	}
	sort.Strings(cellName)
	return
}

//excel导出(数据源为map)
func ExportExcelByMap(titleList []string, data []map[string]interface{}, fileName, sheetName string) (responseFileName string) {

	f := excelize.NewFile()
	f.SetSheetName("Sheet1", sheetName)
	header := make([]string, 0)
	for _, v := range titleList {
		header = append(header, v)
	}
	//表格样式
	rowStyleID, _ := f.NewStyle(`{"font":{"color":"#666666","size":13,"family":"arial"},"alignment":{"vertical":"center","horizontal":"center"}}`)
	_ = f.SetSheetRow(sheetName, "A1", &header)
	_ = f.SetRowHeight(sheetName, 1, 30)
	length := len(titleList)
	headStyle := Letter(length)
	var lastRow string
	var widthRow string
	for k, v := range headStyle {
		if k == length-1 {
			lastRow = fmt.Sprintf("%s1", v)
			widthRow = v
		}
	}
	if err := f.SetColWidth(sheetName, "A", widthRow, 30); err != nil {
		fmt.Println(err)
	}
	rowNum := 1
	for _, value := range data {
		row := make([]interface{}, 0)
		var dataSlice []string
		for key := range value {
			dataSlice = append(dataSlice, key)
		}
		sort.Strings(dataSlice)
		for _, v := range dataSlice {
			if val, ok := value[v]; ok {
				row = append(row, val)
			}
		}
		rowNum++
		if err := f.SetSheetRow(sheetName, fmt.Sprintf("A%d", rowNum), &row); err != nil {
			fmt.Println(err)
		}
		if err := f.SetCellStyle(sheetName, fmt.Sprintf("A%d", rowNum), fmt.Sprintf("%s", lastRow), rowStyleID); err != nil {
			fmt.Println(err)
		}

	}
	//disposition := fmt.Sprintf("attachment; filename=%s-%s.xlsx", url.QueryEscape(fileName), time.Now().Format("20060102150405"))
	//c.Writer.Header().Set("Content-Type", "application/octet-stream")
	//c.Writer.Header().Set("Content-Disposition", disposition)
	//c.Writer.Header().Set("Content-Transfer-Encoding", "binary")
	//c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
	//return f.Write(c.Writer)
	//生成目录
	fileDir := createFile(fileName)
	if fileDir != "" {
		responseFileName = fileDir + fileName + "_" + time.Now().Format("20060102150405") + ".xlsx"
		if err := f.SaveAs(responseFileName); err != nil {
			fmt.Println(err)
		}
		time.Sleep(time.Second * 5)
		fmt.Println("我已经将文件>>", fileName+".xlsx", "成功写入！！！")
		CheckTask.Delete(fileName)
	} else {
		fmt.Println("无法找对对应的目录!")
	}
	return responseFileName
}

func createFile(taskid string) string {
	fileDir := ""
	fmt.Println("开始生成目录..>")
	var config = helper.Config{}
	err := configor.Load(&config, "config.yml")
	if err != nil {
		panic(err)
	}
	sysType := runtime.GOOS
	if sysType == "linux" {
		fileDir = config.Exportfile + taskid + "/"
	} else if sysType == "windows" {
		fileDir = config.Exportfile + taskid + "\\"
	}
	fmt.Println("本次上次的目录为>>>", fileDir)
	err = os.MkdirAll(fileDir, os.ModePerm) //先去创建文件夹
	if err != nil {
		fmt.Println("创建文件夹失败了>>>", err)
	}
	return fileDir

}

// Letter 遍历a-z
func Letter(length int) []string {
	var str []string
	for i := 0; i < length; i++ {
		str = append(str, string(rune('A'+i)))
	}
	return str
}

// ExportExcelByStruct excel导出(数据源为Struct)
func ExportExcelByStruct(titleList []string, data []interface{}, fileName string, sheetName string) {
	f := excelize.NewFile()
	f.SetSheetName("Sheet1", sheetName)
	header := make([]string, 0)
	for _, v := range titleList {
		header = append(header, v)
	}
	rowStyleID, _ := f.NewStyle(`{"font":{"color":"#666666","size":13,"family":"arial"},"alignment":{"vertical":"center","horizontal":"center"}}`)
	_ = f.SetSheetRow(sheetName, "A1", &header)
	_ = f.SetRowHeight("Sheet1", 1, 30)
	length := len(titleList)
	headStyle := Letter(length)
	var lastRow string
	var widthRow string
	for k, v := range headStyle {
		if k == length-1 {
			lastRow = fmt.Sprintf("%s1", v)
			widthRow = v
		}
	}
	if err := f.SetColWidth(sheetName, "A", widthRow, 30); err != nil {

	}
	rowNum := 1
	for _, v := range data {
		t := reflect.TypeOf(v)
		value := reflect.ValueOf(v)
		row := make([]interface{}, 0)
		for l := 0; l < t.NumField(); l++ {
			val := value.Field(l).Interface()
			row = append(row, val)
		}
		rowNum++
		err := f.SetSheetRow(sheetName, "A"+strconv.Itoa(rowNum), &row)
		_ = f.SetCellStyle(sheetName, fmt.Sprintf("A%d", rowNum), fmt.Sprintf("%s", lastRow), rowStyleID)
		if err != nil {

		}
	}
	if err := f.SaveAs(fileName + ".xlsx"); err != nil {
		fmt.Println(err)
	}
}
