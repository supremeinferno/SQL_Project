
{{define "content"}}
<h3 class="mb-4">Admin Dashboard</h3>

<div class="row">
  <div class="col-md-4">
    <div class="card p-4 mb-4">
      <h5 class="mb-3">Add Student</h5>
      <form method="POST" action="/add-student">
        <input class="form-control mb-2" name="name" placeholder="Name" required>
        <input class="form-control mb-2" name="roll" placeholder="Roll No" required>
        <input class="form-control mb-2" name="room" placeholder="Room No" required>
        <input class="form-control mb-2" name="username" placeholder="Username" required>
        <input type="password" class="form-control mb-3" name="password" placeholder="Password" required>
        <button class="btn btn-primary w-100">Add Student</button>
      </form>
    </div>
  </div>

  
  <div class="col-md-8">
    <div class="card p-4 mb-4">
      <h5 class="mb-3">Complaints</h5>
      <table class="table table-striped table-hover">
        <thead class="table-dark">
          <tr>
            <th>Student</th>
            <th>Roll</th>
            <th>Room</th>
            <th>Title</th>
            <th>Status</th>
          </tr>
        </thead>
        <tbody>
        {{range .Complaints}}
          <tr>
            <td>{{.StudentName}}</td>
            <td>{{.RollNo}}</td>
            <td>{{.RoomNo}}</td>
            <td>{{.Title}}</td>
            <td>
              <span class="badge {{if eq .Status "Resolved"}}bg-success{{else}}bg-warning{{end}}">
                {{.Status}}
              </span>
            </td>
          </tr>
        {{end}}
        </tbody>
      </table>
    </div>
  </div>
</div>

{{end}} 
