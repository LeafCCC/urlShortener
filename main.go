package main

import (
	"fmt"
	"net/http"
)

var store *UrlStore

const AddForm = `
<html><body>
<form method="POST" action="/add">

URL: <input type="text" name="url">

<input type="submit" value="Add">

</form>
</body><html>
`

func Add(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue("url")
	//如果url为空 则返回一个表单供用户输入url
	//一个函数实现两个处理效果 很有意思！
	if url == "" {
		fmt.Fprintf(w, AddForm)
		return
	}

	//原路径不是http开头的 进行添加
	if url[:8] != "https://" || url[:7] != "http://" {
		url = "https://" + url
	}

	key := store.Put(url)
	fmt.Fprintf(w, "http://localhost:8080/%s", key)
}

func Redirect(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len("/"):]
	if path == "" {
		http.NotFound(w, r)
		return
	}
	res := store.Get(path)

	if res == "" {
		http.NotFound(w, r)
		return
	}

	//！！！这个重定向 大坑中的大坑 记得要http://开头 不然是相对路径
	//在Add函数中进行了bug修复
	http.Redirect(w, r, res, http.StatusFound)
}

func main() {
	store = NewUrlStore()

	http.HandleFunc("/", Redirect)

	http.HandleFunc("/add", Add)

	http.ListenAndServe(":8080", nil)

}
