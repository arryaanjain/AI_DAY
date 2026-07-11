package ai
import "context"
type ImageRequest struct{Prompt, SourceAssetURL string}
type ImageResult struct{URL, ProviderRequestID string}
// Provider isolates OpenAI and future image backends from business workflows.
type Provider interface{GenerateImage(context.Context,ImageRequest)(ImageResult,error); ModerateText(context.Context,string)(bool,error)}
