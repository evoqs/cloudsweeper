package scheduler

import (
	logger "cloudsweep/logging"
	"cloudsweep/model"
	"cloudsweep/runner"
	"cloudsweep/storage"
	"fmt"
	"sync"
	"time"

	"github.com/go-co-op/gocron"
)

var pipelineSchedulerStore *PipeLineSchedulerStore

func init() {
	// TODO: Implement the lazy construction
	pipelineSchedulerStore = &PipeLineSchedulerStore{
		pipelineSchedulers: make(map[string]*PipeLineScheduler),
		scheduleMux:        sync.Mutex{},
		log:                logger.NewDefaultLogger(),
	}
}

type PipeLineSchedulerStore struct {
	pipelineSchedulers map[string]*PipeLineScheduler
	scheduleMux        sync.Mutex
	log                logger.Logger
}

func (pss *PipeLineSchedulerStore) CreatePipelineScheduler(name string, log logger.Logger, pipelineOperator *storage.PipeLineOperator) (*PipeLineScheduler, error) {
	pss.scheduleMux.Lock()
	defer pss.scheduleMux.Unlock()
	log.Infof("This is working \n")
	// Check if a scheduler with the given ID already exists
	if _, exists := pss.pipelineSchedulers[name]; exists {
		return nil, fmt.Errorf("Scheduler with name " + name + " already exists.")
	}
	//ctx, cancel := context.WithCancel(context.Background())
	pss.pipelineSchedulers[name] = &PipeLineScheduler{
		Scheduler: &Scheduler{
			Name:      name,
			taskMap:   make(map[string]Task),
			scheduler: gocron.NewScheduler(time.UTC),
			log:       log,
		},
		pipelineOperator: pipelineOperator,
		log:              log,
		/*ctx:              ctx,
		cancel:           cancel,
		controlChannel:   make(chan struct{}),
		pipelineChannel:  make(chan model.PipeLine),*/
	}
	pss.log.Infof("Created PipeLineScheduler with name: " + name)
	return pss.pipelineSchedulers[name], nil
}

// Gets the Scheduler. Creates a new Scheduler if not present.
func (pss *PipeLineSchedulerStore) GetPipelineScheduler(name string) (*PipeLineScheduler, error) {
	pss.scheduleMux.Lock()
	defer pss.scheduleMux.Unlock()

	// Check if a scheduler with the given ID already exists
	if existingScheduler, exists := pss.pipelineSchedulers[name]; exists {
		return existingScheduler, nil
	}
	return nil, fmt.Errorf("PipeLineScheduler with name " + name + " doesn't exist.")
}

type PipeLineScheduler struct {
	*Scheduler
	log              logger.Logger
	runMux           sync.Mutex
	pipelineOperator *storage.PipeLineOperator
	// Fields for context and control channel
	/*ctx             context.Context
	cancel          context.CancelFunc
	controlChannel  chan struct{}
	pipelineChannel chan model.PipeLine*/
}

func (pw *PipeLineScheduler) ScheduleAllPipelines() error {
	pw.runMux.Lock()
	defer pw.runMux.Unlock()
	//Get the pipeline json
	pipelines, err := pw.pipelineOperator.GetAllPipeLines()
	if err != nil {
		pw.log.Errorf("Problem in fetching all pipelines from DB: %s", err)
		return fmt.Errorf("Problem in fetching all pipelines from DB: %s", err)
	}
	pw.log.Infof("Total number of pipelines: %d", len(pipelines))

	for _, pipeline := range pipelines {
		if pipeline.Enabled {
			//pw.pipelineChannel <- pipeline
			pw.AddPipelineSchedule(pipeline)
		}
	}
	return nil
}

func (pw *PipeLineScheduler) AddPipelineSchedule(pipeline model.PipeLine) error {
	pw.log.Debugf("Adding PipelineSchedule for pipeline: " + pipeline.PipeLineID.String())
	cronExpression := pw.getScheduleFromPipeline(pipeline)
	err := pw.AddCron(pipeline.PipeLineID.Hex(), cronExpression, func() { runner.ValidateAndRunPipeline(pipeline.PipeLineID.Hex()) })
	//err := pw.AddCron(pipeline.PipeLineID.Hex(), cronExpression, func() { pw.log.Warnf("Scheduled call") })
	if err != nil {
		pw.log.Infof("PipelineScheduler - "+pw.Name+": Error while Adding cron for pipeline: %s Error: %s", pipeline.PipeLineName, err)
		return err
	}
	pw.log.Debugf("Added PipelineSchedule for pipeline: " + pipeline.PipeLineID.String())
	return nil
}

func (pw *PipeLineScheduler) UpdatePipelineSchedule(pipeline model.PipeLine) error {
	cronExpression := pw.getScheduleFromPipeline(pipeline)
	err := pw.UpdateCron(pipeline.PipeLineID.Hex(), cronExpression, func() { runner.ValidateAndRunPipeline(pipeline.PipeLineID.Hex()) })
	if err != nil {
		pw.log.Infof("PipelineScheduler - "+pw.Name+": Error while Updating cron for pipeline: %s Error: %s", pipeline.PipeLineName, err)
		return err
	}
	return nil
}

func (pw *PipeLineScheduler) DeletePipelineSchedule(pipelineId string) error {
	pw.log.Debugf("Deleting PipelineSchedule for pipeline: " + pipelineId)
	for _, task := range pw.ListCron() {
		pw.log.Debugf("Task = %s", task)
	}
	err := pw.DeleteCron(pipelineId)
	if err != nil {
		pw.log.Infof("PipelineScheduler - "+pw.Name+": Error while Deleting cron for pipeline: %s Error: %s", pipelineId, err)
		return err
	}
	pw.log.Debugf("Deleted PipelineSchedule for pipeline: " + pipelineId)
	return nil
}

/*func (pw *PipeLineScheduler) startPipelineScheduler() {
	pw.startScheduler()
	go func() {
		pw.log.Infof("PipelineScheduler - " + pw.Name + ": Started")
		defer pw.log.Infof("PipelineScheduler - " + pw.Name + ": Stopped")

		// Use a select statement to handle start and stop signals
		for {
			select {
			case <-pw.controlChannel:
				pw.log.Infof("PipelineScheduler - " + pw.Name + ": Received stop signal")
				return // Stop the scheduler goroutine
			case pipeline := <-pw.pipelineChannel:
				pw.log.Infof("PipelineScheduler - "+pw.Name+": Received pipeline: %+v\n", pipeline)
				// Call the schedulePipeline function with the received pipeline
				err := pw.schedulePipeline(pipeline)
				if err != nil {
					pw.log.Errorf("PipelineScheduler - "+pw.Name+": Error while adding cron for pipeline: %s Error: %s", pipeline.PipeLineName, err)
				}
			}
		}
	}()
}*/

/*func (pw *PipeLineScheduler) stopPipelineScheduler() {
	pw.cancel()                     // Cancel the context to signal the scheduler to stop
	pw.controlChannel <- struct{}{} // Send a stop signal to the control channel
	pw.stopScheduler()
}*/

// Constructs the schedule from the pipeline model object
// TODO: Should this be part of model struct?
func (pw *PipeLineScheduler) getScheduleFromPipeline(pipeline model.PipeLine) string {
	cronExpression := pipeline.Schedule.Minute + " " + pipeline.Schedule.Hour + " " + pipeline.Schedule.DayOfMonth + " " + pipeline.Schedule.Month + " " +
		pipeline.Schedule.DayOfWeek
	return cronExpression
}

func startPipelineScheduler(name string) *PipeLineScheduler {
	opr := storage.GetDefaultDBOperators()
	var pipelineScheduler *PipeLineScheduler
	pipelineScheduler, err := pipelineSchedulerStore.GetPipelineScheduler(name)
	if err != nil {
		logger.NewDefaultLogger().Infof("Creating PipelineScheduler: " + name)
		pipelineScheduler, err = pipelineSchedulerStore.CreatePipelineScheduler(name, logger.NewDefaultLogger(), &opr.PipeLineOperator)
	}
	logger.NewDefaultLogger().Infof("Starting PipelineScheduler: " + name)
	pipelineScheduler.StartScheduler()
	logger.NewDefaultLogger().Infof("Started PipelineScheduler: " + name)
	return pipelineSchedulerStore.pipelineSchedulers[name]
}

func GetPipelineScheduler(name string) *PipeLineScheduler {
	return pipelineSchedulerStore.pipelineSchedulers[name]
}

// Entry Point. Starts the scheduler and schedules all the pipelines
func StartDefaultPipelineScheduler() *PipeLineScheduler {
	return startPipelineScheduler("DefaultPipelineScheduler")
}

func GetDefaultPipelineScheduler() *PipeLineScheduler {
	return pipelineSchedulerStore.pipelineSchedulers["DefaultPipelineScheduler"]
}

/*func SchedulePipeline(pipeline model.PipeLine) error {
	//err := GetPipelineScheduler("DefaultPipelineScheduler").schedulePipeline(pipeline)
	//return err
	if pipeline.Enabled {
		GetPipelineScheduler("DefaultPipelineScheduler").pipelineChannel <- pipeline
	}
	return nil
}

func StopPipelineScheduler() error {
	GetPipelineScheduler("DefaultPipelineScheduler").stopPipelineScheduler()
	return nil
}*/
