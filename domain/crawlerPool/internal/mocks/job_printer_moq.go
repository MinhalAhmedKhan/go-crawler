// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mocks

import (
	"monzoCrawler/domain/crawlerPool"
	"monzoCrawler/domain/models"
	"sync"
)

// Ensure, that JobPrinterMock does implement crawlerPool.JobPrinter.
// If this is not the case, regenerate this file with moq.
var _ crawlerPool.JobPrinter = &JobPrinterMock{}

// JobPrinterMock is a mock implementation of crawlerPool.JobPrinter.
//
// 	func TestSomethingThatUsesJobPrinter(t *testing.T) {
//
// 		// make and configure a mocked crawlerPool.JobPrinter
// 		mockedJobPrinter := &JobPrinterMock{
// 			PrintFunc: func(job models.CrawlJob)  {
// 				panic("mock out the Print method")
// 			},
// 		}
//
// 		// use mockedJobPrinter in code that requires crawlerPool.JobPrinter
// 		// and then make assertions.
//
// 	}
type JobPrinterMock struct {
	// PrintFunc mocks the Print method.
	PrintFunc func(job models.CrawlJob)

	// calls tracks calls to the methods.
	calls struct {
		// Print holds details about calls to the Print method.
		Print []struct {
			// Job is the job argument value.
			Job models.CrawlJob
		}
	}
	lockPrint sync.RWMutex
}

// Print calls PrintFunc.
func (mock *JobPrinterMock) Print(job models.CrawlJob) {
	if mock.PrintFunc == nil {
		panic("JobPrinterMock.PrintFunc: method is nil but JobPrinter.Print was just called")
	}
	callInfo := struct {
		Job models.CrawlJob
	}{
		Job: job,
	}
	mock.lockPrint.Lock()
	mock.calls.Print = append(mock.calls.Print, callInfo)
	mock.lockPrint.Unlock()
	mock.PrintFunc(job)
}

// PrintCalls gets all the calls that were made to Print.
// Check the length with:
//     len(mockedJobPrinter.PrintCalls())
func (mock *JobPrinterMock) PrintCalls() []struct {
	Job models.CrawlJob
} {
	var calls []struct {
		Job models.CrawlJob
	}
	mock.lockPrint.RLock()
	calls = mock.calls.Print
	mock.lockPrint.RUnlock()
	return calls
}