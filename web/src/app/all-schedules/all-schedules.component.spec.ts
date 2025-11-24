import { ComponentFixture, TestBed } from '@angular/core/testing';

import { AllSchedulesComponent } from './all-schedules.component';

describe('AllSchedulesComponent', () => {
  let component: AllSchedulesComponent;
  let fixture: ComponentFixture<AllSchedulesComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [AllSchedulesComponent]
    })
    .compileComponents();

    fixture = TestBed.createComponent(AllSchedulesComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
