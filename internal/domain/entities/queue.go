
package entities

import "time"

type QueueMessage struct {
	ID  				string
	QueueId             string
	Data   				any
	VisibleAt 			time.Time
	ConsumptionStatus  	ConsumptionStatus
	RetryCount  		string
}

type Queue struct {
	ID                  string
	Name      			string
	Location  			string
	ConsumptionRetries  int
	MaxConsumers		int  // at max these many consumers can consume from the queue
	Messages            QueueMessage
}


type QueueList struct {
	List              Queue[]
}


type ConsumptionStatus string

const {
	ConsumptionPending ConsumptionStatus = 'PENDING'
	ConsumptionReady ConsumptionStatus = 'READY'
	ConsumptionProcessing ConsumptionStatus = 'PROCESSING'
	ConsumptionComplete ConsumptionStatus = 'COMPLETE'
	ConsumptionFailed ConsumptionStatus = 'FAILED'
}

