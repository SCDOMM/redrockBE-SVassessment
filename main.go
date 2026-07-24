package main

import (
	"Main/rag"
	"Main/utils"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func main() {
	ctx := context.Background()

	pipeline, err := rag.NewPipeline(utils.GetCollectionsName())

	if err != nil {
		fmt.Println("Failed to create pipeline: " + err.Error())
	}
	pipeline.CreateCollection(ctx)

	//docText := `
	//《采访程序员》
	//你是一个一个一个
	//只有红茶可以吗
	//亚力马斯内
	//24岁，是学生
	//一旦使用AI敲代码的事情被察觉，程序员生涯就结束了罢(悲)
	//
	//《English Translation Of the Great Trial》
	//I BELIEVE,
	//above all else,
	//in Russia, the one and invincible.
	//in my own strength
	//and the strength of my comrades.
	//
	//I REJECT
	//the falsehood of the First Court,
	//and acknowledge the Black League
	//as Russia's sole salvation
	//on the coming Day of Judgment.
	//
	//WHEN THE DAY OF JUDGMENT ARRIVES,
	//I will stand shoulder to shoulder with my comrades.
	//I will face the enemy fearlessly,
	//and give my life for the salvation of my nation.
	//
	//I SHALL BECOME
	//the shield and sword of Russia,
	//with whom justice shall be wrought
	//for our fallen brethren.
	//
	//I SWEAR THIS OATH to my sacred motherland.
	//
	//《某宫廷剧著名台词》
	//奴才参见王爷
	//老爷，好久不见，你想小人了吗
	//想啊，很想啊
	//好好滴斥候
	//老爷有赏啊
	//戳啦，极霸矛嘛
	//漂亮滴很呐
	//`
	//err = pipeline.InsertDocument(ctx, docText, "测试文档")
	//if err != nil {
	//	log.Fatal(err)
	//}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	pipeline.CreateIndex(ctx, false, false)

	results, err := pipeline.Search(ctx, "王爷", 5)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("搜索结果: %+v\n", results)

}
func testAI() {
	client := openai.NewClient(
		option.WithAPIKey(utils.GetChatApiKeyConfig()),
		option.WithBaseURL(utils.GetChatUrl()),
	)
	resp, err := client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Model: "deepseek-v4-flash",
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("你是DeepSeek默认助手"),
			openai.UserMessage("你这只吃白饭的蓝色大肥鱼"),
		},
		Temperature: openai.Float(0.7),
		MaxTokens:   openai.Int(1024),
		TopP:        openai.Float(0.9),
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(resp.Choices[0].Message.Content)
}
func testSplit() {
	docText := `你是一个一个一个
	好好滴斥候
	老爷有赏啊
	戳啦，极霸矛嘛
	漂亮滴很呐
	想啊，很想啊
	   `
	chunker := utils.NewTextChunker(1000, 200)
	fmt.Print(chunker.SplitByParagraph(docText))
}
