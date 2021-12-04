
每日自动调用Tscancode, Web 查看结果

## 1 Run

run Source/BackEnd/bin/EveryDayScan using root

or

run Source/BackEnd/src/build.sh

We can get results by visiting `http://yourip:port`

## 2 See Source/BackEnd/bin/config.ini

[Project]
Path: Scan path of project
Dirs: Project's dir

[ScanTime]
Invoke Tscancode everyday except NotScanDay at 'Hour' before 'Minute'

[ScanTool]
Path:  Where is Tscancode 
Param = -q -j1
OutputDir = ../../../Tscancode/output/
ResultFile = result
NotScanDay = 0,6

[CodeVcs]
User: Project's Version Control user, not root
Cmd = svn up

[LogToFile]
Open = 0
Path = ../../../everyday.log

[Port]
Port = 9100

## 3 Example
We can get daily results and all results

![image](https://github.com/wangxiaobai-dd/EveryDayTscan_WebShow/blob/4e2bfc156826fa73ee78bcd75b3fe562199d40f0/images/example.png)
