package main 

import (
	"fmt"
	"github.com/dragonmaster101/go_chat/chat"
)

var API_TOKEN string = "{your_hugging_face_api_token}";
var API_URL string = "https://api-inference.huggingface.co/models/facebook/blenderbot-400M-distill";

func main(){

	convo := chat.Conversation{};
	convo.Init(nil);
	convo.Auth(API_TOKEN , API_URL);
	
	fmt.Println(convo.QueryTest("testing!!"));
	convo.SaveLog();
}
