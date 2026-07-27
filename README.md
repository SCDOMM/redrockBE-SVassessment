REDROCK后端部门暑假考核项目

开发者：SCDOMM

# 一.简要介绍

这是REDROCK后端部门的暑假考核项目，要求是从deepwiki、agent、agent框架中任选其一进行开发，本项目选择仿写deepwiki，为避嫌将项目名改为了"dickwiki"(显然很不优雅)

由于时间问题，本项目很多功能没有优化或者加强，比如对文件重新索引等等。

已经实现的功能有：

- 给一个 GitHub 公开仓库地址,系统能拉取仓库并完成摄取
- 摄取时会默认跳过 `.git/`、`vendor/`、`node_modules/`、二进制、超大文件等，如果用户在API中传入自定义的 include/exclude 规则也可以
- 每段代码块都能追溯到来源文件(相对于仓库的路径)、语言等
- 能基于代码内容做问答，如果有需要可以在答案里指明引用了哪些文件
- 摄取异步进行:提交仓库立即返回任务 ID，之后能凭 ID 查到进度，并且能凭ID取消任务进度
- 索引持久化到本地，重启服务后不用重新摄取(因为是存milvus里的)
- 如果仓库更新，可以通过milvus里存的hash信息来更新、删除、新增对应的文件，而不是全部覆盖，即增量更新(实验中，不确定有无致命bug)
- 部署时可以根据需要通过配置文件变动使用的LLM
- 通过milvus的Partition Key实现了分仓库管理，加快查询速度


**1.技术栈**

- golang和它的动物朋友们
- Hertz框架和它的必要依赖
- MySQL用于储存对话上下文
- yaml/v3用于解析配置文件，防止硬编码或者api_key被神秘人拉走
- GORM用于和数据库交互
- go-git/v5用于从仓库中拉取代码
- milvus用于储存向量
- milvus Go SDK/v2用于和milvs进行交互
- openai-go/v3用于和AI交互
- crypto/sha256用于处理git仓库中的文件哈希实现增量更新

**2.注意事项**

本项目没有配套前端，请使用postman等进行测试运行。

预先请安装milvus数据库，可以选用docker compose或者官网的配置文件进行安装。

# 二.项目结构

1.api层负责直接与前端交互，接收JSON，发送JSON，序列化，反序列化等等，顺便把错误全部打包，然后发送出去，其将数据传入sv层，或者从sv层传出数据。

包含有api_chat.go负责与对话相关的接口，api_ingest.go负责与摄取仓库相关的接口

2.dao层负责直接和数据库交互，不处理任何业务逻辑，只要调用MySQL搞CRUD就行，其将结果和错误返回给sv层。

dao_init.go用于初始化数据库，dao_chat.go用于对对话的上下文进行CRUD，dao_clean.go会把超时三个月无人对话的记录统统删掉。

3.model层负责储存数据模型和DTO。

其包含有model_chat.go,model_ingest.go,model_universal.go分别负责封装聊天相关的请求响应，封装摄取仓库相关的请求响应，封装通用请求响应的结构体

4.router层负责注册路由，把对应的地址与api层的函数绑定。

其中包含有router_init.go文件，仅负责注册相关路由。

5.sv层负责处理业务逻辑，比如判断前后端的值是不是空值，或者把dao层的错误打包送走，直接进行摄取/和LLM打交道，又或者是把DTO转化成model，方便dao层处理数据。

其包含有sv_chat.go，sv_ingest.go，sv_init.go分别负责聊天，摄取，初始化结构体相关的逻辑。除此之外，其内部还有三个包：
- intake包负责进行摄取，内部包含有IntakeTask结构体，所有函数围绕其展开。
- intake_init.go会拉取远程仓库，并且生成IntakeTask结构体;intake_manager.go负责管理IntakeTask，保证IntakeTask的并发安全，并且允许动态取消任务;intake_process.go负责运行IntakeTask，其会从拉取的仓库中一个个摄取所有文件，并且进行增量更新，然后塞给rag包进行处理

- llm包负责和LLM聊天，里面只有一个llm_chat.go负责和AI聊天和增添上下文

- rag包负责文本增强检索(RAG)，内部包含有PipeLine结构体，所有函数围绕其展开 
- rag_embedding.go会调用embedding模型将代码/文档向量化；rag_init.go负责生成PipeLine结构体；rag_insert.go负责调用rag_embedding.go中的函数将代码/文档向量化随后存入milvus中；rag_milvus.go负责对milvus进行操作，如新建集合，创建索引，获取文件哈希值等；rag_search.go负责进行搜索，将问题等转为向量然后与milvus中的储备进行比对。 utils层即工具层，负责分割文本/代码，获取yaml文件中的配置，获取哈希值，用雪花算法生成ID等

├─ config.yaml # 配置文件，如数据库、LLM 等参数

├─ go.mod # Go 模块定义

├─ main.go # 程序入口，初始化各层并启动服务

├─ api # 接口层：接收/返回 JSON，序列化/反序列化，错误统一处理

│ ├─ api_chat.go # 对话相关接口

│ └─ api_ingest.go # 摄入仓库相关接口

├─ dao # 数据访问层：直接与 MySQL 交互，执行 CRUD，不处理业务逻辑

│ ├─ dao_init.go # 数据库初始化（连接、建表等）

│ ├─ dao_clean.go # 清除过期对话

│ └─ dao_chat.go # 聊天上下文相关 CRUD 操作

├─ model # 数据模型和 DTO

│ ├─ model_chat.go # 聊天相关的请求/响应结构体

│ ├─ model_ingest.go # 摄入仓库相关的请求/响应结构体

│ └─ model_universal.go # 通用请求/响应结构体

├─ router # 路由层：注册 API 路径与处理函数绑定

│ └─ router_init.go # 路由注册函数

├─ sv # 业务逻辑层：处理校验、错误包装、调度服务等

│ ├─ sv_chat.go # 聊天业务逻辑

│ ├─ sv_ingest.go # 摄入业务逻辑

│ ├─ sv_init.go # 初始化结构体等逻辑

│ ├─ intake/ # 摄入子包：拉取远程仓库、管理/执行摄入任务

│ │ ├─ intake_init.go # 拉取远程仓库，生成 IntakeTask 结构体

│ │ ├─ intake_manager.go # 管理 IntakeTask 的并发安全与动态取消

│ │ └─ intake_process.go # 执行 IntakeTask：逐文件摄入、增量更新，交给 rag 处理

│ ├─ llm/ # LLM 子包：与 AI 聊天及上下文管理

│ │ └─ llm_chat.go # 与 AI 聊天并维护上下文

│ └─ rag/ # RAG 子包：向量化、检索

│ ├─ rag_embedding.go # 调用 Embedding 模型将代码/文档向量化

│ ├─ rag_init.go # 生成 PipeLine 结构体

│ ├─ rag_insert.go # 调用嵌入函数，将向量存入 Milvus

│ ├─ rag_milvus.go # 操作 Milvus：新建集合、创建索引、获取文件哈希等

│ └─ rag_search.go # 搜索：将问题向量化并与 Milvus 中的向量比对

└─ utils # 工具层：文本分割、配置读取、哈希计算、ID 生成等

   ├─ utils_chunk.go # 文本/代码分割

   ├─ utils_config.go # 读取 YAML 配置

   ├─ utils_hash.go # 获取哈希值

   ├─ utils_milvus.go # 查看milvus里有什么字段等

   └─ utils_snowflake.go # 雪花算法生成唯一 ID


# 三.功能与使用

注意！请事先配置好config.yaml文件！
~~~yaml
AiConfig:  
  chat_api_key:  使用聊天模型的APIkey  
  chat_url:  使用聊天模型的API链接，这里默认是  
  chat_model: 使用的聊天模型  
  chat_context: 保留多少轮对话  
  embed_api_key: 使用embed模型的APIkey  
  embed_url: 使用embed模型的API链接  
  embed_model: 使用的embed模型，推荐使用千问  
  embed_dim: 使用embed模型转换文档/代码时用的维数。注意！修改后需要重建一遍集合！  
MilvusConfig:  
  collections_name:  使用的milvus集合名  
GeneralConfig:  
  machine_id:  用雪花算法生成ID时必要的machine_id,随便填个数字即可  
NetworkConfig:  
  network_host: 服务地址  
  network_port: 服务端口  
MySQLConfig:  
  user_name: MySQL用户名  
  password:  MySQL密码  
  host: MySQL地址  
  port: MySQL端口  
  database_name: 在MySQL中储存AI对话上下文时用到的数据库名  
  extra_config: 额外配置  
DefaultFilterPath: 读取仓库时默认忽略的文件夹名，如果用户不传入，就使用它  
  [".git",".idea","vendor","bin","obj","node_modules"]  
DefaultExclude: 读取仓库时默认忽略的后缀名，如果用户不传入，就使用它  
  [".png", ".jpg", ".jpeg", ".gif", ".bmp", ".ico", ".svg", ".webp", ".mp3", ".wav", ".mp4", ".avi", ".pdf", ".doc", ".docx", ".xls", ".xlsx", ".zip", ".tar", ".gz", ".bz2", ".exe", ".dll", ".so", ".dylib", ".bin", ".obj", ".lib", ".ttf", ".otf", ".woff"]  
DefaultInclude: 读取仓库时默认包括的后缀名，如果用户不传入，就使用它  
  [".go", ".py", ".js", ".ts", ".java", ".c", ".cpp", ".h", ".hpp", ".cs", ".rb", ".php", ".swift", ".kt", ".scala", ".rs", ".sh", ".bash", ".zsh", ".yaml", ".yml", ".json", ".xml", ".toml", ".md", ".txt", ".html", ".css", ".scss", ".sql", ".r", ".m", ".mm"]
~~~

### 3.1新建摄取任务

POST：baseUrl/api/ingest/new

请求体的JSON格式如下，其中filter_path和include_patterns，exclude_patterns属性可传入空值，程序将调用yaml文件中配置的默认值进行代替：

~~~JSON
{

  "repo_url": "https://github.com/SCDOMM/DailyHomeWork",

  "filter_path": [

    ".git",".idea","vendor","bin","obj","node_modules"...

  ],

  "include_patterns": [

    ".go", ".py", ".js", ".ts", ".java", ".c", ".cpp", ".h", ".hpp"...

  ],

  "exclude_patterns": [

    ".png", ".jpg", ".jpeg", ".gif", ".bmp", ".ico", ".svg", ".webp"...

  ]

}
~~~

您将获得任务ID的返回值，您可以凭借任务ID来实时查询摄取的进度。

~~~JSON
{

    "Status": "200",

    "Info": "success",

    "Data": {

        "task_id": 7487403717545693185

    }

}
~~~

### 3.2检查摄取任务状态

GET：baseUrl/api/ingest/:id/status

不需要请求体，但您需要在:id的Params中传入任务ID来获取相关信息。

您将在返回的JSON中看到任务进度和任务中出现的报错(部分报错不会阻止摄取进程)

~~~JSON
{

    "Status": "200",

    "Info": "success",

    "Data": {

        "task_id": 7487403717545693185,

        "status": "running",

        "progress": 0,

        "total_file": 124,

        "indexed_file": 0,

        "error": [

            "创建索引失败: rag:fail to override index vector index cannot be dropped on loaded collection: 467959600146284853: invalid parameter"

        ]

    }

}
~~~

如果这个仓库在milvus中加载过，只是更新一些小东西，那么程序会自动缩减相关的流程，只会摄取更新后的文件。比如这里，我往测试仓库中添加了两个txt文件，这里的total_file就变为了2。

~~~JSON
{

    "Status": "200",

    "Info": "success",

    "Data": {

        "task_id": 7487406072257318913,

        "status": "completed",

        "progress": 1,

        "total_file": 2,

        "indexed_file": 2,

        "error": null

    }

}
~~~

### 3.3取消摄取任务

POST：baseUrl/api/ingest/delete

请求体的JSON格式如下，您只需要传入task_id即可取消相关的任务

~~~JSON
{

    "task_id": 7487406072257318913,

}
~~~

响应体将平平无奇。

~~~JSON
{

    "Status": "200",

    "Info": "success",

    "Data": ""

}
~~~

### 3.4新建对话

POST baseUrl/api/chat/new

在相关的仓库摄取完毕后，您就可以和LLM进行对话了，可以在对话中询问仓库相关的问题

请求体需要填入仓库Url，问题，以及过滤路径，过滤路径将从给LLM的选择片段中排除掉您选择的哪些文件或者文件夹。

您可以不传入过滤路径，默认将是没有过滤路径

~~~JSON
{

    "repo_url": "https://github.com/SCDOMM/DailyHomeWork",

    "question": "推测下这个仓库用的是什么语言",

    "filter_path": [

      ".git","java/"...

    ]

}
~~~

您将得到以下响应结果：

~~~JSON
{

    "Status": "200",

    "Info": "success",

    "Data": {

        "chat_id": 7487409317818142721,

        "answer": "根据提供的文档片段，仓库中的代码均为Java语言（如`.java`文件），因此推测该仓库主要使用**Java**。"

    }

}
~~~

### 3.5继续对话

POST baseUrl/api/chat/continue

请保存好您的chat_id，您将在这用到它。

请求体需要传入chat_id和问题，这样您就能继承之前的上下文了。

在这里您也可以一样传入filter_path来过滤掉一些令人讨厌的路径，您也可以选择不传入它。

~~~JSON
{
    "chat_id": 7487409317818142721,

    "question": "好！很有精神！帮我看看仓库中涉及socket和TCP协议的代码有哪些，在哪个文件，在什么相对路径上！"
    
    "filter_path": [

      ".git","java/"...

    ]
}
~~~

您将获得以下响应体：

~~~JSON
{

    "Status": "200",

    "Info": "success",

    "Data": {

        "answer": "根据提供的文档片段，仓库中涉及Socket和TCP协议的代码主要位于以下文件（均使用Java的`java.net.Socket`和`java.net.ServerSocket`实现）：\n\n1. **Library_V8.0\\src\\server\\net\\Server.java**  \n   相对路径：`JavaShortSemester\\Library_V8.0\\src\\server\\net\\Server.java`  \n   开头有`import java.net.*;`，并创建`ServerSocket ss = new ServerSocket(8080);`，通过`Socket s = ss.accept()`接受客户端连接，使用`DataInputStream`和`DataOutputStream`进行TCP通信。\n\n2. **Library_V8.0\\src\\client\\net\\Client.java**  \n   相对路径：`JavaShortSemester\\Library_V8.0\\src\\client\\net\\Client.java`  \n   代码中创建`Socket socket = new Socket(\"127.0.0.1\", 8080);`，并通过`DataInputStream`和`DataOutputStream`进行TCP通信。\n\n3. **Library_V9.0\\src\\client\\net\\Client.java**  \n   相对路径：`JavaShortSemester\\Library_V9.0\\src\\client\\net\\Client.java`  \n   结构和V8.0的Client相同，同样使用Socket连接`127.0.0.1:8080`。\n\n4. **Library_V9.0\\src\\server\\net\\Server.java**  \n   相对路径：`JavaShortSemester\\Library_V9.0\\src\\server\\net\\Server.java`  \n   类似V8.0的Server，使用`ServerSocket`监听`8080`端口，接受TCP连接。\n\n5. **Library_V10.0\\src\\client\\net\\Client.java**  \n   相对路径：`JavaShortSemester\\Library_V10.0\\src\\client\\net\\Client.java`  \n   与前述Client一致，同样基于Socket实现TCP通信。\n\n以上文件均明确使用`ServerSocket`、`Socket`、`DataInputStream`和`DataOutputStream`，符合TCP协议的Socket编程模式。"

    }

}
~~~

### 3.6查看对话

GET baseUrl/api/chat/:id/get

不需要请求体，但您需要在:id的Params中传入对话ID来获取相关信息。

您将在返回的JSON中看到这段对话的一切情况

~~~JSON
{

    "Status": "200",

    "Info": "success",

    "Data": {

        "ChatId": 7487417229806407681,

        "RepoUrl": "https://github.com/SCDOMM/DailyHomeWork",

        "History": [

            {

                "content": "推测下这个仓库用的是什么语言",

                "role": "user"

            },

            {

                "content": "根据提供的文档片段，仓库中主要包含Java源文件（如`.java`文件），因此该仓库使用的是**Java**语言。",

                "role": "assistant"

            },

            {

                "content": "好！很有精神！帮我看看仓库中涉及socket和TCP协议的代码有哪些，在哪个文件，在什么相对路径上！",

                "role": "user"

            },

            {

                "content": "根据提供的文档片段，仓库中涉及Socket和TCP协议的代码出现在以下文件中（均位于 `JavaShortSemester` 目录下）：\n\n1. **服务器端**：\n   - `JavaShortSemester\\Library_V8.0\\src\\server\\net\\Server.java`（片段1）\n   - `JavaShortSemester\\Library_V9.0\\src\\server\\net\\Server.java`（片段5）\n\n2. **客户端**：\n   - `JavaShortSemester\\Library_V8.0\\src\\client\\net\\Client.java`（片段2）\n   - `JavaShortSemester\\Library_V9.0\\src\\client\\net\\Client.java`（片段4）\n   - `JavaShortSemester\\Library_V10.0\\src\\client\\net\\Client.java`（片段3）\n\n这些文件均使用了 `java.net.ServerSocket`、`java.net.Socket`、`java.io.DataInputStream` 和 `java.io.DataOutputStream` 实现基于 TCP 协议的通信。",

                "role": "assistant"

            }

        ],

        "created_at": "2026-07-27T16:02:57.071+08:00",

        "updated_at": "2026-07-27T16:03:24.345+08:00",

        "deleted_at": null

    }

}
~~~

### 3.7删除对话

POST baseUrl/api/chat/delete

您可以通过这个API来删除某个对话，如果不删除，在没人访问的情况下其将在三个月后自动删除。

请求体只需要传入对应的chat_id即可。

~~~JSON
{

    "chat_id": 7487409317818142721

}
~~~

您将获得平平无奇的响应。

~~~JSON
{

    "Status": "200",

    "Info": "success",

    "Data": null

}
~~~

# 四.本地运行指南

### 4.1按寻思之力运行法

什么Docker compose？什么容器化一键部署？都是洋人的歪门邪道！

如果您不需要一键部署或者对docker有意见，可以直接clone仓库，然后启动Goland或者VS摁造运行。

但您仍然需要用docker部署milvus数据库。

您需要配置好config文件，往其中传入您的apiKey，选择您使用的模型，配置好MySQL的milvus的端口和地址，随后即可启动程序。

稍后，您可以通过ApiFox或者Postman来进行测试。

### 4.2Docker容器化部署

您需要配置好仓库目录下的Docker-compose.yml文件中的MySQL密码，MySQL数据库名，以及端口，milvus容器名等信息，其必须与config.yaml文件中的配置相同。

您需要配置好config文件，往其中传入您的apiKey，选择您使用的模型，配置好MySQL的milvus的端口和地址，保证其和Docker-compose.yml文件中的公共字段尽量一致。

随后您就可以切换到仓库克隆的目录，用`docker compose up -d --build`命令一键部署了。

稍后，您可以通过ApiFox或者Postman来进行测试。

# 五.问题和取舍

考虑到时间问题，我实在是懒得写前端了，没有前端的话流式输出又没什么意义，因此不做考虑。

对于模块树和目录树，这毕竟是个小项目，让AI把整个项目读一遍感觉没必要，因此不做考虑。

对于关键词-向量混合检索，比较复杂且可能改变项目结构，因此不做考虑。

当前仍然存在有很多问题，比如很难追踪问题来源，没有前端，没有RPC搞微服务，没有Redis分流等等，但对于个人，小规模使用和考核项目来说已经足够了。
