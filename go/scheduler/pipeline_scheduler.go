package scheduler

import (
	api "cloudsweep/api"
	logger "cloudsweep/logging"
	"cloudsweep/model"
	storage "cloudsweep/storage"
	"errors"
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
	}
}

type PipeLineSchedulerStore struct {
	pipelineSchedulers map[string]*PipeLineScheduler
	scheduleMux        sync.Mutex
	log                logger.Logger
}

func (pss *PipeLineSchedulerStore) CreatePipelineScheduler(name string, log logger.Logger) (*PipeLineScheduler, error) {
	pss.scheduleMux.Lock()
	defer pss.scheduleMux.Unlock()
	// Check if a scheduler with the given ID already exists
	if _, exists := pss.pipelineSchedulers[name]; exists {
		return nil, fmt.Errorf("Scheduler with name " + name + " already exists.")
	}
	pss.pipelineSchedulers[name] = &PipeLineScheduler{
		Scheduler: &Scheduler{
			Name:      name,
			taskMap:   make(map[string]Task),
			scheduler: gocron.NewScheduler(time.UTC),
			log:       log,
		},
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
	logger.Logger
	runMux sync.Mutex
}

func (pw *PipeLineScheduler) scheduleAllPipelines() error {
	pw.runMux.Lock()
	defer pw.runMux.Unlock()
	//Get the pipeline json
	pipelines, err := storage.GetAllPipelines() //TODO: Waiting for Bibin
	if err != nil {
		return fmt.Errorf("Problem in fetching all pipelines from DB: %s", err)
	}
	pw.log.Infof("Total number of pipelines: %d", len(pipelines))

	var allErrors error
	for _, pipeline := range pipelines {
		err := pw.schedulePipeline(pipeline)
		if err != nil {
			pw.Errorf("Error while adding cron for pipeline: %s Error: %s", pipeline.PipeLineName, err)
			allErrors = errors.New(allErrors.Error() + " \n" + err.Error())
		}
	}
	return allErrors
}

func (pw *PipeLineScheduler) schedulePipeline(pipeline model.PipeLine) error {
	cronExpression := pw.getScheduleFromPipeline(pipeline)
	err := pw.AddCron(pipeline.PipeLineID.Hex(), cronExpression, func() { api.RunPolicy(pipeline.PipeLineID.Hex()) })
	if err != nil {
		pw.Infof("Error while adding cron for pipeline: %s Error: %s", pipeline.PipeLineName, err)
		return err
	}
	return nil
}

// Constructs the schedule from the pipeline model object
// TODO: Should this be part of model struct?
func (pw *PipeLineScheduler) getScheduleFromPipeline(pipeline model.PipeLine) string {
	cronExpression := pipeline.Schedule.Minute + " " + pipeline.Schedule.Hour + " " + pipeline.Schedule.DayOfMonth + " " + pipeline.Schedule.Month + " " +
		pipeline.Schedule.DayOfWeek
	return cronExpression
}

// This is the entry point for pipeline scheduler. Creates the main default scheduler.
func startPipelineScheduler(name string, loggerObject logger.Logger) *PipeLineScheduler {
	var pipelineScheduler *PipeLineScheduler
	pipelineScheduler, err := pipelineSchedulerStore.GetPipelineScheduler(name)
	if err != nil {
		pipelineScheduler, err = pipelineSchedulerStore.CreatePipelineScheduler("DefaultPipelineScheduler", loggerObject)
	}
	pipelineScheduler.startScheduler()
	logger.NewDefaultLogger().Infof("Started PipelineScheduler: " + name)
	return pipelineSchedulerStore.pipelineSchedulers[name]
}

func GetPipelineScheduler(name string) *PipeLineScheduler {
	return pipelineSchedulerStore.pipelineSchedulers[name]
}

// Entry Point. Starts the scheduler and schedules all the pipelines
func StartPipelineScheduler() *PipeLineScheduler {
	defaultPipelineScheduler := startPipelineScheduler("DefaultPipelineScheduler", logger.NewDefaultLogger())
	defaultPipelineScheduler.scheduleAllPipelines()
	return defaultPipelineScheduler
}
