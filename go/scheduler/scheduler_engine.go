// Package scheduler provides the scheduling capabilities and pipeline scheduling functions
package scheduler

import (
	logger "cloudsweep/logging"
	"fmt"
	"sync"
	"time"

	"github.com/go-co-op/gocron"
)

type SchedulerStore struct {
	Schedulers  map[string]*Scheduler
	ScheduleMux sync.Mutex
	Log         logger.Logger
}

func (ss *SchedulerStore) CreateScheduler(name string, log logger.Logger) (*Scheduler, error) {
	ss.ScheduleMux.Lock()
	defer ss.ScheduleMux.Unlock()
	// Check if a scheduler with the given ID already exists
	if _, exists := ss.Schedulers[name]; exists {
		return nil, fmt.Errorf("Scheduler with name " + name + " already exists.")
	}
	ss.Schedulers[name] = &Scheduler{Name: name,
		taskMap:   make(map[string]Task),
		scheduler: gocron.NewScheduler(time.UTC),
		log:       log,
	}
	log.Infof("Created new scheduler with name: " + name)
	return ss.Schedulers[name], nil
}

// Gets the Scheduler. Creates a new Scheduler if not present.
func (ss *SchedulerStore) GetScheduler(name string) (*Scheduler, error) {
	ss.ScheduleMux.Lock()
	defer ss.ScheduleMux.Unlock()

	// Check if a scheduler with the given ID already exists
	if existingScheduler, exists := ss.Schedulers[name]; exists {
		return existingScheduler, nil
	}
	return nil, fmt.Errorf("Scheduler with name " + name + " doesn't exist.")
}

func (ss *SchedulerStore) DeleteScheduler(name string) error {
	ss.ScheduleMux.Lock()
	defer ss.ScheduleMux.Unlock()

	// Check if a scheduler with the given name exists
	if _, exists := ss.Schedulers[name]; exists {
		// Stop and remove the scheduler
		ss.Schedulers[name].scheduler.Stop()
		delete(ss.Schedulers, name)
		ss.Log.Infof("Deleted scheduler with name: " + name)
		return nil
	}

	return fmt.Errorf("Scheduler with name " + name + " doesn't exist.")
}

type Scheduler struct {
	Name       string
	scheduler  *gocron.Scheduler
	taskMap    map[string]Task
	taskMapMux sync.Mutex
	log        logger.Logger
}

type Task struct {
	Id  string
	job *gocron.Job
}

/*
*  This method can be called any number of times, without having any additional effect
*  because, s.scheduler.StartXXX() will be ignored if it is already running
 */
func (s *Scheduler) StartScheduler() {
	s.log.Infof("Start the Scheduler id: " + s.Name)
	go s.scheduler.StartAsync()
}

func (s *Scheduler) StopScheduler() {
	s.log.Infof("Stop the Scheduler id: " + s.Name)
	go s.scheduler.Stop()
}

/* Adds the Cron job as per provided cron expression. Id can be any uuid for tracking
 * If a cron for the give Id already exists, then new cron is not added, error is returned
 */
func (s *Scheduler) AddCron(id string, cronExpression string, jobFun interface{}, params ...interface{}) error {
	//fmt.Printf("AddCron -> %s\n", cronExpression)
	s.taskMapMux.Lock()
	defer s.taskMapMux.Unlock()

	if _, exists := s.taskMap[id]; exists {
		return fmt.Errorf("Cron %s already exists. Skipping the add Cron.", id)
	}
	job, err := s.scheduler.Cron(cronExpression).Tag(id).Do(jobFun, params...)
	s.taskMap[id] = Task{Id: id, job: job}
	s.log.Debugf("Successfully added the cron with Id: %s with Scheduler: %s with Cron: %s", id, s.Name, cronExpression)
	return err
}

// Updates the Cron Expression for the
func (s *Scheduler) UpdateCron(id string, cronExpression string, jobFun interface{}, params ...interface{}) error {
	if err := s.DeleteCron(id); err != nil {
		return err
	}
	return s.AddCron(id, cronExpression, jobFun, params...)
}

func (s *Scheduler) DeleteCron(id string) error {
	job := s.getJobByID(id)
	s.taskMapMux.Lock()
	defer s.taskMapMux.Unlock()
	if job == nil {
		delete(s.taskMap, id)
		return fmt.Errorf("Unable to delete Cron Job for pipeline: " + id + ". Reason: CronJob not found.")
	}
	s.scheduler.RemoveByID(job)
	delete(s.taskMap, id)
	s.log.Debugf("Successfully deleted the cron with Id: %s with Scheduler: %s", id, s.Name)
	return nil
}

func (s *Scheduler) ListCron() []Task {
	s.taskMapMux.Lock()
	defer s.taskMapMux.Unlock()

	var tasksList []Task
	for _, task := range s.taskMap {
		tasksList = append(tasksList, task)
	}
	return tasksList
}

func (s *Scheduler) getJobByID(id string) *gocron.Job {
	s.taskMapMux.Lock()
	defer s.taskMapMux.Unlock()

	// Iterate through the taskMap to find the Task with the matching id
	for _, task := range s.taskMap {
		if task.Id == id {
			return task.job
		}
	}
	return nil
}
