package main


//
//问题描述：
//
//实现一个程序计算并打印输入的目录下所有文件的总数和总大小(以GB计算)。完成之后你将熟悉select、WaitGroup、ioutil的用法。
//
//
//
//要点：
//
//并发读取文件(夹)信息。
//限制开启的goroutines的最大数量。
//运行时每隔500ms打印当前已经统计的文件数和总大小（使用命令行参数指定此功能是否启用）。
//
//
//拓展：
//
//在执行中在有外部输入时退出程序。
//
//
//实现：

import (
"flag"
"fmt"
"io/ioutil"
"os"
"path/filepath"
"sync"
"time"
)

var verbose = flag.Bool("v", false, "show verbose progress messages")
//semaphore 就是信号量.这里面设置50.表示到50就堵塞他,别让cpu开太多县城.
var sema = make(chan struct{}, 50)
var done = make(chan struct{})

func dirents(dir string) []os.FileInfo{
	select {
	case sema <- struct{}{}: //
	case <- done:
		return nil

	}
	// 函数运行完sema再出来一个.
	defer func() {<- sema}()
	// ReadDir reads the directory named by dirname and returns
	// a list of directory entries sorted by filename.
	//返回 fileinfo这个类的一个数组.这个类有size属性.
	entries, err := ioutil.ReadDir(dir)

	if err != nil{
		fmt.Fprintf(os.Stderr, "du1: %v\n", err)
		return nil
	}

	return entries

}

func walkDir(dir string, n *sync.WaitGroup, fileSizes chan<- int64){
	defer n.Done()

	if cancelled(){
		return
	}

	for _, entry := range dirents(dir){
		if entry.IsDir(){
			n.Add(1)
			subdir := filepath.Join(dir, entry.Name())
			go walkDir(subdir, n, fileSizes)
		} else {
			fileSizes <- entry.Size()
		}
	}
}

func printDiskUsage(nfiles, nbytes int64){
	fmt.Printf("%d files %.1f GB\n", nfiles, float64(nbytes/1e9))
}

func cancelled() bool{
	select {
	case <- done:
		return true
	default:
		return false
	}
}

func main() {

	a:=struct{}{}
	fmt.Print(a)//用fmt.Print什么都能打印.


	flag.Parse()
//下面我们跑E盘.
	roots :=[]string{"E:\\"}

	var tick <-chan time.Time

	if *verbose{
		tick = time.Tick(500 * time.Millisecond)
	}

//输入为空就表示当前目录. 在go里面项目中的.表示项目的根目录.这个代码跑完正好是5个文件. 一个main和其他4个 goland自动生成的文件.
	if len(roots) == 0{
		roots = []string{"."}
	}


	fileSizes := make(chan int64)
	var nfiles, nbytes int64

	var n sync.WaitGroup

	for _, root := range roots{
		n.Add(1)
		go walkDir(root, &n, fileSizes)
	}


	//最终函数会进入这个判断,来退出所有子携程.
	go func() {
		n.Wait()
		close(fileSizes) //关闭filesizes这个channel
	}()



//这个代码是用来,手动关闭的,也就是控制台如果输入信息,那么就关闭done信号.这样就会让其他部分
//检测done时候报错,从而强行禁止程序继续跑.
	go func() {
		os.Stdin.Read(make([]byte, 1))//比如我们这时候再控制台输入1回车,那么就会立即停止程序.
		close(done)
	}()

loop:
	for {
		select {


		case <-done:
			for range fileSizes{
				//
			}
		case size, ok := <- fileSizes:
			if !ok {
				break loop
			}

			nfiles++
			nbytes += size
//下面这个行是 每一次tick了就打印一次当前读取的结果.
		case <- tick:
			printDiskUsage(nfiles, nbytes)









		}
	}

	printDiskUsage(nfiles, nbytes)
}