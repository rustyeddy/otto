package mqtt

import (
	"context"
	"errors"
	"testing"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/rustyeddy/otto/messenger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeToken struct {
	waitTimeoutResult bool
	err               error
	waitCalls         int
	waitTimeoutCalls  int
	done              chan struct{}
}

func newFakeToken(waitTimeoutResult bool, err error) *fakeToken {
	ch := make(chan struct{})
	close(ch)
	return &fakeToken{
		waitTimeoutResult: waitTimeoutResult,
		err:               err,
		done:              ch,
	}
}

func (t *fakeToken) Wait() bool {
	t.waitCalls++
	return true
}

func (t *fakeToken) WaitTimeout(time.Duration) bool {
	t.waitTimeoutCalls++
	return t.waitTimeoutResult
}

func (t *fakeToken) Done() <-chan struct{} { return t.done }
func (t *fakeToken) Error() error          { return t.err }

type fakeClient struct {
	connectToken     paho.Token
	publishToken     paho.Token
	subscribeToken   paho.Token
	unsubscribeToken paho.Token

	published      []publishArgs
	subscriptions  []subscriptionArgs
	unsubscribed   []string
	connectedState bool
}

type publishArgs struct {
	topic   string
	qos     byte
	retain  bool
	payload interface{}
}

type subscriptionArgs struct {
	topic   string
	qos     byte
	handler paho.MessageHandler
}

func (c *fakeClient) IsConnected() bool      { return c.connectedState }
func (c *fakeClient) IsConnectionOpen() bool { return c.connectedState }
func (c *fakeClient) Connect() paho.Token    { return c.connectToken }
func (c *fakeClient) Disconnect(uint)        {}

func (c *fakeClient) Publish(topic string, qos byte, retained bool, payload interface{}) paho.Token {
	c.published = append(c.published, publishArgs{topic: topic, qos: qos, retain: retained, payload: payload})
	return c.publishToken
}

func (c *fakeClient) Subscribe(topic string, qos byte, callback paho.MessageHandler) paho.Token {
	c.subscriptions = append(c.subscriptions, subscriptionArgs{topic: topic, qos: qos, handler: callback})
	return c.subscribeToken
}

func (c *fakeClient) SubscribeMultiple(map[string]byte, paho.MessageHandler) paho.Token {
	return newFakeToken(true, nil)
}

func (c *fakeClient) Unsubscribe(topics ...string) paho.Token {
	c.unsubscribed = append(c.unsubscribed, topics...)
	return c.unsubscribeToken
}

func (c *fakeClient) AddRoute(string, paho.MessageHandler) {}
func (c *fakeClient) OptionsReader() paho.ClientOptionsReader {
	return paho.NewOptionsReader(paho.NewClientOptions())
}

type fakeMessage struct {
	topic   string
	payload []byte
	retain  bool
	qos     byte
}

func (m *fakeMessage) Duplicate() bool { return false }
func (m *fakeMessage) Qos() byte       { return m.qos }
func (m *fakeMessage) Retained() bool  { return m.retain }
func (m *fakeMessage) Topic() string   { return m.topic }
func (m *fakeMessage) MessageID() uint16 {
	return 1
}
func (m *fakeMessage) Payload() []byte { return m.payload }
func (m *fakeMessage) Ack()            {}

func TestNewUsesProvidedClientID(t *testing.T) {
	t.Parallel()

	cfg := Config{
		Broker:       "tcp://example:1883",
		ClientID:     "client-1",
		Username:     "user",
		Password:     "pass",
		CleanSession: true,
	}

	p := New(cfg)

	require.NotNil(t, p.opts)
	assert.Equal(t, cfg.ClientID, p.opts.ClientID)
	assert.Equal(t, cfg.Username, p.opts.Username)
	assert.Equal(t, cfg.Password, p.opts.Password)
	assert.Equal(t, cfg.CleanSession, p.opts.CleanSession)
	require.Len(t, p.opts.Servers, 1)
	assert.Equal(t, cfg.Broker, p.opts.Servers[0].String())
}

func TestNewGeneratesClientID(t *testing.T) {
	t.Parallel()

	p := New(Config{Broker: "tcp://example:1883"})
	require.NotNil(t, p.opts)
	assert.Len(t, p.opts.ClientID, len("otto-")+8)
	assert.Contains(t, p.opts.ClientID, "otto-")
}

func TestRandSuffix(t *testing.T) {
	t.Parallel()

	suffix := randSuffix()
	assert.Len(t, suffix, 8)
	for _, r := range suffix {
		assert.True(t, (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'))
	}
}

func TestSetWillWithoutOptions(t *testing.T) {
	t.Parallel()

	p := &Paho{}
	err := p.SetWill("topic", []byte("payload"), true, 1)
	require.Error(t, err)
}

func TestConnectTimeout(t *testing.T) {
	t.Parallel()

	p := &Paho{
		opts: paho.NewClientOptions(),
		c: &fakeClient{
			connectToken: newFakeToken(false, nil),
		},
	}

	err := p.Connect(context.Background())
	require.Error(t, err)
}

func TestConnectReturnsTokenError(t *testing.T) {
	t.Parallel()

	p := &Paho{
		opts: paho.NewClientOptions(),
		c: &fakeClient{
			connectToken: newFakeToken(true, errors.New("connect failed")),
		},
	}

	err := p.Connect(context.Background())
	require.Error(t, err)
}

func TestPublishQoS0DoesNotWait(t *testing.T) {
	t.Parallel()

	token := newFakeToken(true, nil)
	client := &fakeClient{publishToken: token}
	p := &Paho{c: client}

	err := p.Publish(context.Background(), "topic", []byte("payload"), false, 0)
	require.NoError(t, err)
	assert.Equal(t, 0, token.waitTimeoutCalls)
	require.Len(t, client.published, 1)
	assert.Equal(t, "topic", client.published[0].topic)
}

func TestPublishQoS1Waits(t *testing.T) {
	t.Parallel()

	token := newFakeToken(true, nil)
	p := &Paho{c: &fakeClient{publishToken: token}}

	err := p.Publish(context.Background(), "topic", []byte("payload"), false, 1)
	require.NoError(t, err)
	assert.Equal(t, 1, token.waitTimeoutCalls)
}

func TestPublishTimeout(t *testing.T) {
	t.Parallel()

	token := newFakeToken(false, nil)
	p := &Paho{c: &fakeClient{publishToken: token}}

	err := p.Publish(context.Background(), "topic", []byte("payload"), false, 1)
	require.Error(t, err)
}

func TestSubscribeSuccessAndUnsubscribe(t *testing.T) {
	t.Parallel()

	subToken := newFakeToken(true, nil)
	unsubToken := newFakeToken(true, nil)
	client := &fakeClient{
		subscribeToken:   subToken,
		unsubscribeToken: unsubToken,
	}
	p := &Paho{c: client}

	got := make(chan messenger.Message, 1)
	unsub, err := p.Subscribe(context.Background(), "topic", 1, func(m messenger.Message) {
		got <- m
	})
	require.NoError(t, err)
	require.NotNil(t, unsub)
	require.Len(t, client.subscriptions, 1)

	handler := client.subscriptions[0].handler
	handler(client, &fakeMessage{
		topic:   "topic",
		payload: []byte("payload"),
		retain:  true,
		qos:     1,
	})

	select {
	case msg := <-got:
		assert.Equal(t, "topic", msg.Topic)
		assert.Equal(t, []byte("payload"), msg.Payload)
		assert.True(t, msg.Retain)
		assert.Equal(t, byte(1), msg.QoS)
	default:
		require.Fail(t, "expected handler to be called")
	}

	err = unsub()
	require.NoError(t, err)
	assert.Equal(t, 1, unsubToken.waitTimeoutCalls)
	assert.Equal(t, []string{"topic"}, client.unsubscribed)
}

func TestSubscribeTimeout(t *testing.T) {
	t.Parallel()

	client := &fakeClient{subscribeToken: newFakeToken(false, nil)}
	p := &Paho{c: client}

	_, err := p.Subscribe(context.Background(), "topic", 1, func(messenger.Message) {})
	require.Error(t, err)
}

func TestSubscribeTokenError(t *testing.T) {
	t.Parallel()

	client := &fakeClient{subscribeToken: newFakeToken(true, errors.New("sub failed"))}
	p := &Paho{c: client}

	_, err := p.Subscribe(context.Background(), "topic", 1, func(messenger.Message) {})
	require.Error(t, err)
}
