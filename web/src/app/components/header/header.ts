import { Component } from '@angular/core';
import { GlobalsService } from '../../globals.service';
import { RouterLink } from '@angular/router';

@Component({
  selector: 'app-header',
  standalone: true,
  imports: [RouterLink],
  templateUrl: './header.html',
  styleUrl: './header.scss',
})
export class Header {
  companyName: string;

  constructor(private globalsService: GlobalsService) {
    this.companyName = this.globalsService.companyName; // Access the company name from the service
  }
}
