package chat

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	logger "log"
)

/*
The Data Format of the resulting response for the POST http request in the Conversation.Query method
*/
type Response struct {
	GeneratedText string `json:"generated_text"`
	Conversation  struct {
		GeneratedResponses []string `json:"generated_responses"`
		PastUserInputs     []string `json:"past_user_inputs"`
		Text 				 string `json:"text"`
	} `json:"conversation"`
	Warnings []string `json:"warnings"`
}

/*
The Data Format for the body of the POST http request in the Conversation.Query method
*/
type Body struct {
	PastUserInputs     []string `json:"past_user_inputs"`
	GeneratedResponses []string `json:"generated_responses"`
	Text               string   `json:"text"`
}

/*

<----------------------------------------------------------------------------------------------------------->

Conversation Type Start

The DataType for a Conversation :
provides methods for initializing , querying , saving conversation instances
*/
type Conversation struct {
	ModelUrl   string // Huggingface Model Api url 
	Token      string // Huggingface Token

	UserInputs []string
	BotInputs  []string

	Log        *ConversationLog
	LogPath    string
}

// the data type to initialize the Conversation type with a specific configuration
type ConversationOptions struct{
	
	ModelUrl  	 string
	Token 	 	 *string

	LogFilePath  string

	Name 	 	 string

	Empty 	 	 bool
	Basic 	 	 bool
	Save 	 	 bool
	Load		 bool
}

//  <--------------------------------------------------------------------------------------------->
// Conversation Type Methods


func BasicConversationOption(token , url string) *ConversationOptions{
	options := ConversationOptions {ModelUrl: url , Token: &token , Basic: true};
	return &options;
}

func LoadConversationOption(path string , token *string) *ConversationOptions {
	options := ConversationOptions{Token: token , LogFilePath: path , Load: true};
	return &options;
}

func SaveConversationOption(token , url , name string) *ConversationOptions {
	options := ConversationOptions{Token: &token , ModelUrl: url , Name: name, Save: true};
	return &options;
}

/*
Conversation Data Type
<-<-<-<---------------------------------------------------------------------------->->->->
							<-<-<	INIT METHODS START	>->->
*/

/*
INITIALIZES A CONVERSATION BASED ON THE GIVEN CONFIGURATION

Generally use convo to denote a conversation instance
	convo := chat.Conversation{};

VALID CONFIGURATION TYPES :


EMPTY -> configures an empty conversation with just the capacity sized configured

BASIC -> configures a conversation for just basic use

LOAD  -> configures a conversation with the configuration of the loaded log file i.e "{filename}.chat.json"

SAVE  -> configures a conversation with the given settings and creates a log for the chat


CONFIGURATION REQUIREMENTS :


	EMPTY ->  None 

	BASIC ->  TOKEN , URL 

	LOAD  ->  LOG_FILE_PATH 

	SAVE  ->  TOKEN , URL , NAME


SAMPLE INVOCATIONS :

initializes the conversation with the Empty configuration :
	convo.Init(nil);

initializes the conversation with the Basic configuration :
	convo.Init(&chat.BasicConversationOption(Token , Url));

initializes the conversation with the Load configuration :
	convo.Init(&chat.LoadConversationOption("path/to/{yourFileName}.chat.json"));

initializes the conversation with the Save configuration :
	convo.Init(&chat.SaveConversationOption(Token , Url , "testChatName"));
*/
func (convo *Conversation) Init(options *ConversationOptions){
	switch options {
	case nil:
		convo.initEmpty();
		return
	default:	
	}

	switch options.Empty {
	case true:
		convo.initEmpty();
	default:
	}

	switch options.Basic {
	case true:
		convo.initBasic(*options.Token , options.ModelUrl)
	default:
	}

	switch options.Load {
	case true:
		convo.initFromLogFile(options.LogFilePath , options.Token);
	default:
	}

	switch options.Save {
	case true:
		convo.initWithLog(*options.Token , options.ModelUrl , options.Name)
	default:
	}
}

// Initializes an empty conversation with the capacity to store 8 instances
func (convo *Conversation) initEmpty() {
	convo.UserInputs = make([]string , 0 , 8);
	convo.BotInputs = make([]string , 0 , 8);

	convo.ModelUrl = "None";
	convo.Token = "None";
}

func (convo *Conversation) initBasic(token , modelUrl string) {
	convo.initEmpty();
	convo.ModelUrl = modelUrl;
	convo.Token = token;
}

func (convo *Conversation) initWithLog(token , modelUrl , name string) {
	convo.initEmpty();
	convo.ModelUrl = modelUrl;
	convo.Token = token;
	convo.CreateLog(name);
}

/*

Loads a log file i.e "example.chat.json" and decodes it into the ConversationLog data type.It then proceeds to convert
that struct into a Conversation Type.


*/

func (convo *Conversation) initFromLogFile(logFileName string , token *string) {
	convo.initEmpty();

	var log ConversationLog;

	convo.LogPath = logFileName;
	convo.Log = &log;

	logFile , err := os.Open(logFileName);
	if err != nil {
		panic(err);
	}
	defer logFile.Close();

	byteValue , err := ioutil.ReadAll(logFile);
	if err != nil {
		panic(err);
	}

	json.Unmarshal(byteValue , &log);

	convo.initFromLog(&log , token);
}

func (convo *Conversation) initFromLog(log *ConversationLog , token *string){
	
	convo.ModelUrl = log.Model;
	convo.Token = log.Token;

	if log.Safe && token == nil{
		err := fmt.Errorf("in initializing conversation with log File :\n safe mode is enabled\n API TOKEN required , TOKEN : '%v'" , log.Token);
		logger.Fatal(err);
	}

	if log.Safe {
		convo.Token = *token;
	}

	for _ , instance := range log.Instances {
		convo.UserInputs = append(convo.UserInputs, instance.User);
		convo.BotInputs = append(convo.BotInputs , instance.Bot);
	}
}



/*
Conversation Data Type
<-<-<-<---------------------------------------------------------------------------->->->->
							<-<-<	INIT METHODS END  >->->
*/

/*
Configures the initial logging settings and creates a chat log with the given name. 

WARNING : use this method for unlogged chats only. Use the SaveLog method for chats which are already logged 
*/
func (convo *Conversation) CreateLog(name string) {

	fileName := name + ".chat.json";
	
	curDir  , err := os.Getwd();
	if err != nil {
		panic(err);
	}
	
	convo.LogPath = curDir + "/" + fileName;

	file , err := os.Create(fileName);
	if err != nil {
		panic(err);
	}
	defer file.Close();

	log := ConversationLog{};
	options := SafeLogOptions(convo.ModelUrl , convo);
	log.Init(&options);

	convo.Log = &log;

	json.NewEncoder(file).Encode(log);
}

func (convo *Conversation) CreateUnSafeLog(name string) {

	fileName := name + ".chat.json";
	
	curDir  , err := os.Getwd();
	if err != nil {
		panic(err);
	}
	
	convo.LogPath = curDir + "/" + fileName;

	file , err := os.Create(fileName);
	if err != nil {
		panic(err);
	}
	defer file.Close();

	log := ConversationLog{};
	options := UnSafeLogOptions(convo.Token , convo.ModelUrl , convo);
	log.Init(&options);

	convo.Log = &log;

	json.NewEncoder(file).Encode(log);
}




func (convo *Conversation) SaveLog(){
	file , err := os.OpenFile(convo.LogPath , os.O_WRONLY , os.ModeTemporary);
	if err != nil {
		panic(err);
	}
	defer file.Close();

	json.NewEncoder(file).Encode(*convo.Log);
}

/*

Sends a POST request to the hugginface Inference API with the given API_TOKEN and API_MODEL_URL
-> Generates the Body Data and embeds the userInput into it
-> Encodes the Body Data Type into json and then into a byte Buffer
-> Sends request to the url with the body and the Auth header
-> Gets response and decodes the json into the Response Data Type 
-> Returns the generated_response field from the request response as a string
*/
func (convo *Conversation) Query(userInput string) string{

	body := Body{};
	body.GeneratedResponses = convo.BotInputs;
	body.PastUserInputs = convo.UserInputs;
	body.Text = userInput;

	var bytesData bytes.Buffer;
	json.NewEncoder(&bytesData).Encode(&body);

	req , err := http.NewRequest("POST" , convo.ModelUrl , &bytesData);
	if err != nil {
		panic(err);
	}
	req.Header.Set("Authorization" , fmt.Sprintf("Bearer %s" , convo.Token));

	client := &http.Client{};
	response , err := client.Do(req);
	if err != nil {
		panic(err);
	}
	defer response.Body.Close();

	respBody := Response{};
	json.NewDecoder(response.Body).Decode(&respBody);
	
	res := respBody.GeneratedText;


	return res;
}


/*

<----------------------------------------------------------------------------->

Conversation Log Types 

*/



/*
Records a single instance of a given query and its resulting response

E.g
-> User : Hi
-> Bot : Hello How are you?

*/
type ConversationLogInstance struct {
	User string `json:"user"`
	Bot  string `json:"bot"`
}

/*
Records an entire conversation comprising of multiple responses and queries

Uses the type LogInstance to denote a single set of query and response
*/
type ConversationLog struct {
	Safe      bool   `json:"safe"`
	Model     string `json:"model"`
	Token     string `json:"token"`
	Instances []ConversationLogInstance `json:"instances"`
}

type ConversationLogOptions struct {
	Token string 
	ModelUrl string
	conversation *Conversation
	Safe  bool
}

func SafeLogOptions(modelUrl string , convo *Conversation) ConversationLogOptions {
	options := ConversationLogOptions{
		ModelUrl: modelUrl , 
		conversation: convo,
		Safe: true,
	};

	return options;
}

func UnSafeLogOptions(token string , modelUrl string , convo *Conversation) ConversationLogOptions {
	options := ConversationLogOptions{
		Token: token , 
		ModelUrl: modelUrl , 
		conversation: convo,
		Safe: false,
	};

	return options;
}

/*
Initializes a conversation log with the given configuration denoted with the Type ConversationLogOptions
*/
func (log *ConversationLog) Init(options *ConversationLogOptions){
	switch options.Safe {
	case true :
		log.initSafe(options.conversation , options.ModelUrl);
	case false:
		log.initUnSafe(options.conversation , options.Token , options.ModelUrl)
	}
}

func (log *ConversationLog) initUnSafe(convo *Conversation , Token string , Model string) {

	log.Model = Model;
	log.Token = Token;
	log.Safe = false;

	userInputLen := len(convo.UserInputs);
	log.Instances = make([]ConversationLogInstance , userInputLen);

	for i := 0;i < userInputLen;i++ {
		log.Instances[i].User = convo.UserInputs[i];
		log.Instances[i].Bot = convo.BotInputs[i];
	}
} 

func (log *ConversationLog) initSafe(convo *Conversation , Model string){

	log.Model = Model;
	log.Token = "Token is hidden due to safety";
	log.Safe = true;

	userInputLen := len(convo.UserInputs);
	log.Instances = make([]ConversationLogInstance , userInputLen);

	for i := 0;i < userInputLen;i++ {
		log.Instances[i].User = convo.UserInputs[i];
		log.Instances[i].Bot = convo.BotInputs[i];
	}
}

/*

Conversation Log Types End 

<---------------------------------------------------------------------------------------------->

*/

/*

<----------------------------------------------------------------------------------------------->

Random Helper Functions

*/

// terminal input helper function Displays the given string and waits for user input
func Input(prompt string) string{

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("-> " + prompt)
	text, _ := reader.ReadString('\n')
	text = strings.Replace(text, "\n", "", -1)

	return text;
}

// prints formatted output to the terminal
func Print(prompt string){
	fmt.Print("-> " + prompt);
}