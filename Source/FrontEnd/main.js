let dates = []
let dateIndex = -1
let curDate = "无数据"
let pattern = /\[.+\/(.+\/.+)\].*\((Information|Critical|Serious|Warning)\)(.+)/

window.onload = function () {
    getRecordDate(true)
}

let timer = setInterval(function () {
    let hours = new Date().getHours();
    let min = new Date().getMinutes();
    if (hours === '3' && min === '0') {
        getRecordDate(true)
    }
}, 50000)

// date list
function getRecordDate(load) {
    axios.get('/getRecord')
        .then(function (response) {
            dates = response.data
            if (!dates)
                return
            for (var i = 0; i < dates.length; i++) {
                console.log(dates[i])
            }
            if (load && dates.length) {
                containerVue.selectDate(0)
            }
        })
        .catch(function (error) {
            alert(error);
        });
}

let containerVue = new Vue({
    delimiters: ['[[', ']]'],
    el: '#container',
    data: {
        isAICheck: false,
        results: [],
        tmpResults: [],
        aiResults: [],
        selected: '昨日新增',
        date: curDate,
        disable: 0,
        isDump : false,
        dumpVersions:[],
        selectedDumpVersion: 0,
        dumpContent: '',
    },
    computed: {
        buttonText() {
            return this.isAICheck ? 'TS 检查结果' : 'AI 检查结果'
        },
        dumpText() {
            return this.isDump ? 'TS 检查结果' : 'Dump 结果'
        }
    },
    methods: {
        selectDate: function (value) {
            this.disable = 0
            if (!dates || !dates.length)
                return
            console.log("selectdata")

            if (value === 0) {
                dateIndex = dates.length - 1
            } else if (value > 0) {
                if (dateIndex < dates.length - 1)
                    dateIndex += 1
            } else if (value < 0) {
                if (dateIndex > 0)
                    dateIndex -= 1
            }
            if (dateIndex < dates.length) {
                curDate = dates[dateIndex]
                this.date = curDate
                if (this.isDump)
                    this.getDumpDayVersions()
                else if (this.isAICheck)
                    this.getDayAIResult()
                else
                    this.getDayResult()
            }
            if (dateIndex === 0)
                this.disable += 1
            if (dateIndex === dates.length - 1)
                this.disable += 2

        },
        parseAIRawResult(rawData) {
            const revisions = []
            const blocks = rawData.split(/(REVISION:\d+\s+[^\n]+)/g);

            for (let i = 1; i < blocks.length; i += 2) {
                revisions.push({
                    header: blocks[i].trim(),
                    content: this.formatCodeContent(blocks[i + 1].trim())
                })
            }
            return revisions
        },
        formatCodeContent(content) {
            return content.replace(/```cpp\n?([\s\S]*?)```/g, (_, code) => {
                const highlighted = hljs.highlight(code.trim(), {language: 'cpp'}).value;
                return `<pre class="code-block"><code class="hljs cpp">${highlighted}</code></pre>`;
            });
        },
        toggleAICheck: function () {
            this.isAICheck = !this.isAICheck;
            console.log("toggle ai check")
            if (this.isAICheck) {
                this.isDump = false
                console.log(curDate)
                $("#pag").show();
                this.getDayAIResult()
            }
        },
        toggleDump: function () {
            this.isDump = !this.isDump;
            console.log("toggle dump check")
            if (this.isDump) {
                this.isAICheck = false
                $("#pag").show();
                this.getDumpDayVersions()
            }
        },
        getDayAIResult: function () {
            axios.get('/getAIDay', {
                params: {
                    date: curDate
                }
            })
                .then(response => {
                    this.aiResults = this.parseAIRawResult(response.data)
                })
                .catch(function (error) {
                    console.log(error);
                });
        },
        selectDumpVersion(value) {
            console.log(value)
            this.selectedDumpVersion = value.version
            this.getDumpDayResult()
        },
        getDumpDayVersions: function () {
            this.selectedDumpVersion = 0
            this.dumpContent = ""
            axios.get('/getDumpDayVersions', {
                params: {
                    date: curDate
                }
            })
                .then(response => {
                    this.dumpVersions = response.data.versions;
                    console.log(this.dumpVersions)

                    if (this.selectedDumpVersion === 0 && this.dumpVersions && this.dumpVersions.length > 0) {
                        this.selectedDumpVersion = this.dumpVersions[0].version;
                        console.log(this.selectedDumpVersion)
                        this.getDumpDayResult()
                    }
                })
                .catch(function (error) {
                    console.log(error);
                });
        },
        getDumpDayResult: function (){
            axios.get('/getDumpDay', {
                params: {
                    date: curDate,
                    version: this.selectedDumpVersion
                }
            })
                .then(response => {
                    this.dumpContent = response.data;
                })
                .catch(function (error) {
                    this.selectedDumpVersion = 0
                    console.log('获取内容失败:', error);
                    this.dumpContent = '<p class="error">内容加载失败</p>';
                });
        },
        getDayResult: function () {
            axios.get('/getDay', {
                params: {
                    date: curDate
                }
            })
                .then(response => {
                    this.tmpResults = response.data
                    console.log(curDate)
                })
                .catch(function (error) {
                    console.log(error);
                });
        },
        getAllResult: function () {
            axios.get('/getAll')
                .then(response => (this.tmpResults = response.data))
                .catch(function (error) {
                    console.log(error);
                });
        },
        makeSelect: function (value) {
            if (value === 1) {
                this.selected = "昨日新增";
                containerVue.selectDate(0)
                $("#pag").show();
            } else if (value === 2) {
                this.selected = "全部";
                this.getAllResult()
                $("#pag").hide();
            }
        }
    },
    watch: {
        tmpResults: function (value) {
            this.results = []
            if (!value) {
                console.log("tmpResults null")
                return
            }
            console.log("tmplength:", value.length)
            for (var i = 0; i < value.length; ++i) {
                var arr = pattern.exec(value[i])
                var tmp = {}
                if (arr) {
                    tmp["file"] = arr[1]
                    tmp["level"] = arr[2]
                    tmp["info"] = arr[3]
                    switch (arr[2]) {
                        case "Critical":
                            tmp["type"] = 1;
                            break; // 1:red 2:blue 3:yellow 4:green
                        case "Serious":
                            tmp["type"] = 2;
                            break;
                        case "Warning":
                            tmp["type"] = 3;
                            break;
                        case "Information":
                            tmp["type"] = 4;
                            break;
                        default:
                            tmp["type"] = 0;
                    }
                } else {
                    tmp["info"] = value[i]
                }
                this.results.push(tmp)
            }
        }
    },
    updated() {
        // 自动高亮新插入的代码块
        document.querySelectorAll('pre code').forEach((block) => {
            hljs.highlightElement(block);
        });
    }
});

