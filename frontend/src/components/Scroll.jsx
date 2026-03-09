import styled from 'styled-components';
import SimpleBar from 'simplebar-react';
import 'simplebar-react/dist/simplebar.min.css';

const Scroll = styled(SimpleBar)`
  .simplebar-track.simplebar-vertical {
    background-color: #00000000;
    pointer-events: auto;
    z-index: 2;
  }
  .simplebar-scrollbar::before {
    background-color: #5e5e5e !important;
    transition: opacity 0.8s !important;
  }
`;

export default Scroll;

