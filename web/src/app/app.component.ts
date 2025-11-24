import { Component } from '@angular/core';
import { RouterModule, RouterOutlet } from '@angular/router';
import { HomeComponent } from './home/home.component';
import { MySchedulesComponent } from './my-schedules/my-schedules.component';
import { AdminComponent } from './admin/admin.component';
import { Header } from './components/header/header';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [
    RouterModule,
    RouterOutlet,
    Header,
    HomeComponent,
    MySchedulesComponent,
    AdminComponent,
    FormsModule,
  ],
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss'],
})
export class AppComponent {
  title = 'Shift Schedule';
}
