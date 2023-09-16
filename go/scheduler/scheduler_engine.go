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
	schedulers  map[string]*Scheduler
	scheduleMux sync.Mutex
	log         logger.Logger
}

func (ss *SchedulerStore) CreateScheduler(name string, log logger.Logger) (*Scheduler, error) {
	ss.scheduleMux.Lock()
	defer ss.scheduleMux.Unlock()
	// Check if a scheduler with the given ID already exists
	if _, exists := ss.schedulers[name]; exists {
		return nil, fmt.Errorf("Scheduler with name " + name + " already exists.")
	}
	ss.schedulers[name] = &Scheduler{Name: name,
		taskMap:   make(map[string]Task),
		scheduler: gocron.NewScheduler(time.UTC),
		log:       log,
	}
	log.Infof("Created new scheduler with name: " + name)
	return ss.schedulers[name], nil
}

// Gets the Scheduler. Creates a new Scheduler if not present.
func (ss *SchedulerStore) GetScheduler(name string) (*Scheduler, error) {
	ss.scheduleMux.Lock()
	defer ss.scheduleMux.Unlock()

	// Check if a scheduler with the given ID already exists
	if existingScheduler, exists := ss.schedulers[name]; exists {
		return existingScheduler, nil
	}
	return nil, fmt.Errorf("Scheduler with name " + name + " doesn't exist.")
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
func (s *Scheduler) startScheduler() {
	s.log.Infof("Start the Scheduler id: " + s.Name)
	go s.scheduler.StartAsync()
}

func (s *Scheduler) stopScheduler() {
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
	s.log.Debugf("111111111")
	job := s.getJobByID(id)
	s.log.Debugf("222222222")
	s.taskMapMux.Lock()
	defer s.taskMapMux.Unlock()
	if job == nil {
		s.log.Debugf("3333333")
		delete(s.taskMap, id)
		s.log.Debugf("4444444")
		return fmt.Errorf("Unable to delete Cron Job for pipeline: " + id + ". Reason: CronJob not found.")
	}
	s.log.Debugf("5555555555")
	s.scheduler.RemoveByID(job)
	s.log.Debugf("666666666")
	delete(s.taskMap, id)
	s.log.Debugf("777777777")
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
