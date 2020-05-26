package pubsub

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	//"fmt"
)

const (
	PUBLISH = "publish"
	SUBSCRIBE = "subscribe"
	UNSUBSCRIBE = "unsubscribe"
)

//list all clients
type PubSub struct {
	Clients  []Client
	Subscriptions []Subscription
}

type Client struct {
	Id string
	Connection *websocket.Conn
}

type Message struct {
	Action string `json:"action"`
	Topic string `json:"topic"`
	Message json.RawMessage `json:"message"`
}

type Subscription struct {
	Topic string
	Client *Client
}

func(client *Client) Send(message []byte) error {
	return client.Connection.WriteMessage(1, message)

}

func(ps *PubSub)RemoveClient(client Client)*PubSub{
	//first remove all subscriptions by this client
	for index, sub := range ps.Subscriptions{
		if client.Id == sub.Client.Id {
			ps.Subscriptions = append(ps.Subscriptions[:index],ps.Subscriptions[index+1:]...)
		}
	}
	//remove client from the list
	for index, c := range ps.Clients{
		if c.Id == client.Id {
			ps.Clients = append(ps.Clients[:index],ps.Clients[index+1:]...)
		}
	}
	return ps
}

func(ps *PubSub) Publish(topic string, message []byte, excludeClient * Client){
	subscriptions := ps.GetSubscriptions(topic,nil)

	for _, sub := range subscriptions {
		fmt.Printf("Sending to client id %s message is %s",sub.Client.Id, message)
		// send the message to all clients
		sub.Client.Send(message)
	}
}

func(ps *PubSub) GetSubscriptions(topic string, client *Client) []Subscription{

	var subscriptionList []Subscription
	for _,subscription := range ps.Subscriptions{
		//for finding a subscription from client id  and also topic
		if client != nil {
			if subscription.Client.Id == client.Id && subscription.Topic  == topic {
				subscriptionList = append(subscriptionList, subscription)
			}
		}else {
				//for finding a subscription only from topic
				if subscription.Topic == topic {
					subscriptionList =  append(subscriptionList, subscription)
				}
		}
	}
	return subscriptionList
}
func(ps *PubSub) AddClient(client Client) *PubSub{
	ps.Clients = append(ps.Clients, client)

	//fmt.Println("Adding new client to the list.", client.Id, len(ps.Clients))
	payload := []byte("Hello Client with ID: " +client.Id)
	client.Connection.WriteMessage(1,payload)

	return ps
}

func (ps *PubSub) Unsubscribe(client *Client, topic string) *PubSub {

	for index, sub := range  ps.GetSubscriptions(topic, client) {
		if sub.Client.Id == client.Id && sub.Topic == topic{
		//	find the subscription and remove it
			ps.Subscriptions =  append(ps.Subscriptions[:index], ps.Subscriptions[index+1:]...)
		}
	}

}
func (ps *PubSub) Subscribe(client *Client, topic string) *PubSub {

	clientSubs := ps.GetSubscriptions(topic, client)

	if len(clientSubs)>0 {
		return ps
	}

	newSubscription := Subscription{
		Topic: topic,
		Client: client,
	}

	ps.Subscriptions = append(ps.Subscriptions, newSubscription)

	return ps
}

func (ps *PubSub) HandleReceiveMessage(client Client, messageType int, payload []byte)(*PubSub){
	m := Message{}

	err:= json.Unmarshal(payload, &m)
	if err != nil {
		fmt.Println("This is not a correct payload")
		return ps
	}
	//fmt.Println("Client is corrected. Message payload is ",m.Action, string(m.Message), m.Topic)
	switch m.Action {
	// Find all subscriber with same topic and send this client
	case PUBLISH :
		fmt.Println("This is a publish new method.")
		ps.Publish(m.Topic, m.Message, &client)
		break

	// add subscriber to subscriptionlists
	case SUBSCRIBE :
		ps.Subscribe(&client, m.Topic)
		fmt.Println("New subscriber is added with topic",m.Topic, len(ps.Subscriptions))

		break
	case UNSUBSCRIBE :
		fmt.Println("New subscriber is deleted.")
		ps.Unsubscribe(&client, m.Topic)
		break
	default :
		break
	}
	return ps


}

