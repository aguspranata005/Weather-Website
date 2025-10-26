// src/react-mwc.d.ts
import React from 'react';

declare module 'react' {
  namespace JSX {
    interface IntrinsicElements {
      'md-filled-button': React.DetailedHTMLProps<React.HTMLAttributes<HTMLElement>, HTMLElement>;
      'md-outlined-button': React.DetailedHTMLProps<React.HTMLAttributes<HTMLElement>, HTMLElement>;
      'md-outlined-text-field': React.DetailedHTMLProps<React.HTMLAttributes<HTMLInputElement> & { label?: string; value?: string; }, HTMLInputElement>;
      'md-circular-progress': React.DetailedHTMLProps<React.HTMLAttributes<HTMLElement> & { indeterminate?: boolean; }, HTMLElement>;
      'md-list': React.DetailedHTMLProps<React.HTMLAttributes<HTMLElement>, HTMLElement>;
      'md-list-item': React.DetailedHTMLProps<React.HTMLAttributes<HTMLElement> & { headline?: string; }, HTMLElement>;
      'md-divider': React.DetailedHTMLProps<React.HTMLAttributes<HTMLElement>, HTMLElement>;
    }
  }
}