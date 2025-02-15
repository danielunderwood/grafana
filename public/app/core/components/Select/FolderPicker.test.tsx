import { render, screen } from '@testing-library/react';
import React from 'react';
import selectEvent from 'react-select-event';

import { selectors } from '@grafana/e2e-selectors';
import * as api from 'app/features/manage-dashboards/state/actions';

import { DashboardSearchHit } from '../../../features/search/types';

import { FolderPicker, getInitialValues } from './FolderPicker';

describe('FolderPicker', () => {
  it('should render', async () => {
    jest
      .spyOn(api, 'searchFolders')
      .mockResolvedValue([
        { title: 'Dash 1', id: 1 } as DashboardSearchHit,
        { title: 'Dash 2', id: 2 } as DashboardSearchHit,
      ]);

    render(<FolderPicker onChange={jest.fn()} />);
    expect(await screen.findByTestId(selectors.components.FolderPicker.containerV2)).toBeInTheDocument();
  });

  it('Should apply filter to the folders search results', async () => {
    jest
      .spyOn(api, 'searchFolders')
      .mockResolvedValue([
        { title: 'Dash 1', id: 1 } as DashboardSearchHit,
        { title: 'Dash 2', id: 2 } as DashboardSearchHit,
        { title: 'Dash 3', id: 3 } as DashboardSearchHit,
      ]);

    render(<FolderPicker onChange={jest.fn()} filter={(hits) => hits.filter((h) => h.id !== 2)} />);

    const pickerContainer = screen.getByLabelText(selectors.components.FolderPicker.input);
    selectEvent.openMenu(pickerContainer);

    const pickerOptions = await screen.findAllByLabelText('Select option');

    expect(pickerOptions).toHaveLength(2);
    expect(pickerOptions[0]).toHaveTextContent('Dash 1');
    expect(pickerOptions[1]).toHaveTextContent('Dash 3');
  });
});

describe('getInitialValues', () => {
  describe('when called with folderId and title', () => {
    it('then it should return folderId and title', async () => {
      const getFolder = jest.fn().mockResolvedValue({});
      const folder = await getInitialValues({ folderId: 0, folderName: 'Some title', getFolder });

      expect(folder).toEqual({ label: 'Some title', value: 0 });
      expect(getFolder).not.toHaveBeenCalled();
    });
  });

  describe('when called with just a folderId', () => {
    it('then it should call api to retrieve title', async () => {
      const getFolder = jest.fn().mockResolvedValue({ id: 0, title: 'Title from api' });
      const folder = await getInitialValues({ folderId: 0, getFolder });

      expect(folder).toEqual({ label: 'Title from api', value: 0 });
      expect(getFolder).toHaveBeenCalledTimes(1);
      expect(getFolder).toHaveBeenCalledWith(0);
    });
  });

  describe('when called without folderId', () => {
    it('then it should throw an error', async () => {
      const getFolder = jest.fn().mockResolvedValue({});
      await expect(getInitialValues({ getFolder })).rejects.toThrow();
    });
  });
});
