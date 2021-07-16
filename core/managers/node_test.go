package managers

/*func init() {

	InitNode(context.Background(), false)
}

func TestRemoveContainer(t *testing.T) {

	cname := GetContainerName(1)

	RemoveContainer(context.Background(), cname)

	c := SearchDockerContainers(context.Background(), "name", cname, true)
	if len(c) != 0 {
		t.Fatal("ERROR:", t.Name())
	}
}

func TestNewContainer(t *testing.T) {

	cname := GetContainerName(1)

	NewContainer(context.Background(), &inputs.CtrInfo, cname)

	c := SearchDockerContainers(context.Background(), "name", cname, false)
	if len(c) != 1 {
		t.Fatal("ERROR:", t.Name())
	}
}

func TestStopContainer(t *testing.T) {

	cname := GetContainerName(1)

	StopContainer(context.Background(), cname)

	c1 := SearchDockerContainers(context.Background(), "name", cname, true)
	c2 := SearchDockerContainers(context.Background(), "name", cname, false)
	if len(c1) != 1 || len(c2) != 0 {
		t.Fatal("ERROR:", t.Name())
	}
}

func TestStartContainer(t *testing.T) {

	cname := GetContainerName(1)

	StartContainer(context.Background(), cname)

	c := SearchDockerContainers(context.Background(), "name", cname, false)
	if len(c) != 1 {
		t.Fatal("ERROR:", t.Name())
	}
}

func TestRestartContainer(t *testing.T) {

	cname := GetContainerName(1)

	RestartContainer(context.Background(), cname)

	c := SearchDockerContainers(context.Background(), "name", cname, false)
	if len(c) != 1 {
		t.Fatal("ERROR:", t.Name())
	}
}*/
