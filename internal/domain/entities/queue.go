
package entities

import "time"

type QueueMessage struct {
	id  				string
	QueueId             string
	data   				any
	visibleAt 			time.Time
	consumptionStatus  	ConsumptionStatus
	retryCount  		string
}

type Queue struct {
	id                  string
	name      			string
	location  			string
	consumptionRetries  int
	maxConsumers		int  // at max these many consumers can consume from the queue
	messages            QueueMessage
}


type QueueList struct {
	list              Queue[]
}


type ConsumptionStatus string

const {
	ConsumptionPending ConsumptionStatus = 'PENDING'
	ConsumptionReady ConsumptionStatus = 'READY'
	ConsumptionProcessing ConsumptionStatus = 'PROCESSING'
	ConsumptionComplete ConsumptionStatus = 'COMPLETE'
	ConsumptionFailed ConsumptionStatus = 'FAILED'
}

