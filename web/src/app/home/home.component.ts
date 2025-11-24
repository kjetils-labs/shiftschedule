import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';

@Component({
  selector: 'app-home',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterLink],
  templateUrl: './home.component.html',
  styleUrl: './home.component.scss',
})
export class HomeComponent implements OnInit {
  tables = [
    {
      tableName: 'Table 1',
      data: [
        {
          weekNumber: 1,
          assignee: 'John Doe',
          substitute: 'Jane Smith',
          accepted: 'Yes',
          comment: 'No issues',
        },
        {
          weekNumber: 2,
          assignee: 'Alice Brown',
          substitute: 'Bob White',
          accepted: 'No',
          comment: 'Waiting for confirmation',
        },
        {
          weekNumber: 3,
          assignee: 'Charlie Green',
          substitute: 'David Blue',
          accepted: 'Yes',
          comment: 'Confirmed',
        },
      ],
    },
    {
      tableName: 'Table 2',
      data: [
        {
          weekNumber: 1,
          assignee: 'Eva Black',
          substitute: 'Tom Grey',
          accepted: 'Yes',
          comment: 'All good',
        },
        {
          weekNumber: 2,
          assignee: 'Paul Red',
          substitute: 'Sophia Yellow',
          accepted: 'Yes',
          comment: 'On track',
        },
        {
          weekNumber: 3,
          assignee: 'Olivia White',
          substitute: 'Liam Brown',
          accepted: 'No',
          comment: 'Waiting for approval',
        },
        {
          weekNumber: 4,
          assignee: 'James Blue',
          substitute: 'Noah Green',
          accepted: 'Yes',
          comment: 'Confirmed',
        },
      ],
    },
    {
      tableName: 'Table 3',
      data: [
        {
          weekNumber: 1,
          assignee: 'Mia Pink',
          substitute: 'Lucas Orange',
          accepted: 'Yes',
          comment: 'Everything fine',
        },
        {
          weekNumber: 2,
          assignee: 'Ethan Violet',
          substitute: 'Grace Yellow',
          accepted: 'No',
          comment: 'Pending',
        },
        {
          weekNumber: 3,
          assignee: 'Chloe Green',
          substitute: 'Liam Black',
          accepted: 'Yes',
          comment: 'All set',
        },
      ],
    },
  ];
  currentWeek: number = 1; // Default value for current week

  constructor() { }

  ngOnInit(): void {
    // This is where you'd fetch data from your API later
  }
}
