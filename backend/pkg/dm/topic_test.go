package dm

import "testing"

func TestTopicBuilderBuildsSubjectsFromSegments(t *testing.T) {
	got := Topic().
		Org("1001").
		Session("sess_0").
		Message().
		Build()

	if want := "org.1001.session.sess_0.message"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestTopicBuilderBuildsWildcardSubjects(t *testing.T) {
	got := Topic().
		Org("1001").
		Add("worker").
		Wildcard().
		Task().
		Build()

	if want := "org.1001.worker.*.task"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestTopicBuilderCleansSegments(t *testing.T) {
	if got, want := Topic().Add(" org.1001 ", " worker.1 ").Build(), "org_1001.worker_1"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}

	if got, want := Topic().Add("").Build(), "unknown"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}

	if got, want := Topic().Add("worker * >").Build(), "worker____"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestTopicBuilderIsImmutable(t *testing.T) {
	base := Topic().Org("1001")
	task := base.Worker("worker_1").Task()
	stream := base.Session("sess_1").Message().Stream()

	if got, want := base.Build(), "org.1001"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	if got, want := task.Build(), "org.1001.worker.worker_1.task"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	if got, want := stream.Build(), "org.1001.session.sess_1.message.stream"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestTopicBuilderWithSeparator(t *testing.T) {
	got := Topic().
		Org("1001").
		Worker("worker_1").
		Task().
		WithSeparator("_").
		Build()

	if want := "org_1001_worker_worker_1_task"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestTopicBuilderWithUnderscoreSeparator(t *testing.T) {
	got := Topic().
		Org("1001").
		Worker("worker_1").
		Task().
		WithUnderscoreSeparator().
		Build()

	if want := "org_1001_worker_worker_1_task"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestTopicBuilderWithSeparatorIsImmutable(t *testing.T) {
	base := Topic().Org("1001").Worker("worker_1")
	underscore := base.WithSeparator("_").Task()

	if got, want := base.Task().Build(), "org.1001.worker.worker_1.task"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	if got, want := underscore.Build(), "org_1001_worker_worker_1_task"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
