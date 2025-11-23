import { Component } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { Components } from './components/components';
import { Header } from './components/header/header';
import { Home } from './home/home';

@Component({
  selector: 'app-root',
  imports: [RouterOutlet, Components, Header, Home],
  templateUrl: './app.component.html',
  styleUrl: './app.component.scss',
})
export class AppComponent {
  title = 'shiftschedule';
}
