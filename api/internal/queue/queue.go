package queue
import "context"
// Queue is deliberately provider-neutral: Redis can be used in production and
// a database-backed implementation can be used for a minimal deployment.
type Message struct{Topic,JobID,TraceID string;Attempt int}
type Queue interface{Publish(context.Context,Message)error;Receive(context.Context,string)(Message,error);Acknowledge(context.Context,Message)error}
