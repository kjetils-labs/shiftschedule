import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';

@Component({
  selector: 'app-my-schedules',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './my-schedules.component.html',
  styleUrl: './my-schedules.component.scss',
})
export class MySchedulesComponent { }
