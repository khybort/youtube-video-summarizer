package kafka

// Topic names for Kafka
const (
	// Video processing topics
	TopicVideoCreated        = "video.created"
	TopicTranscriptRequested = "video.transcript.requested"
	TopicEmbeddingRequested  = "video.embedding.requested"
	TopicSimilarityRequested = "video.similarity.requested"
	
	// Status topics
	TopicAnalysisCompleted = "video.analysis.completed"
	TopicAnalysisFailed    = "video.analysis.failed"
	
	// Dead letter queue
	TopicDLQ = "video.dlq"
)

// Consumer group names
const (
	GroupTranscriptWorker = "transcript-worker"
	GroupEmbeddingWorker  = "embedding-worker"
	GroupSimilarityWorker = "similarity-worker"
	GroupDLQProcessor    = "dlq-processor"
)

// GetTopicPartitions returns the number of partitions for a topic
// This can be configured per environment
func GetTopicPartitions(topic string) int {
	// Default partition counts - adjust based on throughput needs
	partitions := map[string]int{
		TopicVideoCreated:        3,
		TopicTranscriptRequested: 5, // Higher for I/O intensive work
		TopicEmbeddingRequested:  3,
		TopicSimilarityRequested: 2,
		TopicAnalysisCompleted:   2,
		TopicAnalysisFailed:      1,
		TopicDLQ:                 1,
	}
	
	if count, ok := partitions[topic]; ok {
		return count
	}
	return 1 // Default to 1 partition
}

// GetTopicReplicationFactor returns the replication factor for a topic
func GetTopicReplicationFactor() int {
	// In production, use 3 for high availability
	// In development, use 1
	return 1
}

