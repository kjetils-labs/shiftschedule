import { Injectable } from '@angular/core';

@Injectable({
  providedIn: 'root',
})
export class GlobalsService {
  // You can change this to whatever company name you want
  companyName: string = 'Shift Schedule';

  constructor() { }
}
