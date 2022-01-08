package main 

import (
	"fmt"
	"github.com/dragonmaster101/go_chat/chat"
)

var API_TOKEN string = "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";
var API_URL string = "https://api-inference.huggingface.co/models/facebook/blenderbot-400M-distill";

// func Start(){
// 	convo := chat.Conversation{};
// 	convo.Init();
// 	for {
// 		input := chat.Input("User : ");
// 		reply := chat.Query(input , &convo);
// 		out := fmt.Sprintf("Bot : %v\n" , reply);
// 		chat.Print(out);
// 	}
// }

func main(){

	convo := chat.Conversation{};
	convo.Init(chat.SaveConversationOption(API_TOKEN , API_URL , "test"));

	fmt.Println(convo.Query("hi"));
	convo.SaveLog();
}