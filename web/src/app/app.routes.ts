import { Routes } from '@angular/router';
import { HomeComponent } from './home/home.component';
import { MySchedulesComponent } from './my-schedules/my-schedules.component';
import { AdminComponent } from './admin/admin.component';
import { AllSchedulesComponent } from './all-schedules/all-schedules.component';

export const routes: Routes = [
  { path: '', redirectTo: '/home', pathMatch: 'full' },
  { path: 'home', component: HomeComponent },
  { path: 'schedules/mine', component: MySchedulesComponent },
  { path: 'admin', component: AdminComponent },
  { path: 'schedules/all', component: AllSchedulesComponent },
];
